package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/iam"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/user"
	"github.com/aromancev/confa/user/session"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Token struct {
	Token     string `json:"token"`
	ExpiresIn int32  `json:"expiresIn"`
}

type Session struct {
	EmailToken string `json:"emailToken"`
}

type Login struct {
	Address string `json:"address"`
}

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Handler struct {
	router http.Handler
}

func NewHandler(baseURL string, secretKey *auth.SecretKey, publicKey *auth.PublicKey, sessions *session.CRUD, users *user.CRUD, producer Producer, tubeEmail string) *Handler {
	r := http.NewServeMux()

	r.HandleFunc("/health", ok)
	r.Handle(
		"/token",
		fetchToken(publicKey, secretKey, sessions),
	)
	r.Handle(
		"/session",
		createSession(publicKey, secretKey, users, sessions),
	)
	r.Handle(
		"/login",
		login(baseURL, secretKey, producer, tubeEmail),
	)
	r.Handle(
		"/logout",
		logout(sessions),
	)

	return &Handler{
		router: r,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, traceID := trace.Ctx(r.Context())
	w.Header().Set("Trace-Id", traceID)

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeHTTP panic")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	lw := newResponseWriter(w)
	r = r.WithContext(ctx)
	h.router.ServeHTTP(lw, r)

	lw.Event(ctx, r).Msg("Web served")
}

func ok(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, code: http.StatusOK}
}

func (w *responseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Event(ctx context.Context, r *http.Request) *zerolog.Event {
	var event *zerolog.Event
	if w.code >= http.StatusInternalServerError {
		event = log.Ctx(ctx).Error()
	} else {
		event = log.Ctx(ctx).Info()
	}
	return event.Str("method", r.Method).Int("code", w.code).Str("url", r.URL.String())
}

func fetchToken(publicKey *auth.PublicKey, secretKey *auth.SecretKey, sessions *session.CRUD) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		authCtx := auth.NewHTTPContext(w, r)

		claims := func() *auth.APIClaims {
			if s, err := sessions.Fetch(ctx, authCtx.Session()); err == nil {
				return auth.NewAPIClaims(s.Owner, auth.AccountUser)
			}
			var c auth.APIClaims
			if err := publicKey.Verify(authCtx.GuestClaims(), &c); err == nil {
				return &c
			}
			return auth.NewGuesAPIClaims(uuid.New())
		}()

		access, err := secretKey.Sign(claims)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to sign API token.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if claims.Account == auth.AccountGuest {
			authCtx.SetGuestClaims(access)
		} else {
			authCtx.ResetGuestClaims()
		}

		_ = json.NewEncoder(w).Encode(Token{
			Token:     access,
			ExpiresIn: int32(claims.ExpiresIn().Seconds()),
		})
	}
}

func createSession(publicKey *auth.PublicKey, secretKey *auth.SecretKey, users *user.CRUD, sessions *session.CRUD) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var sessionRequest Session
		err := json.NewDecoder(r.Body).Decode(&sessionRequest)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Failed to unmarshal session.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var claims auth.EmailClaims
		err = publicKey.Verify(sessionRequest.EmailToken, &claims)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Email verification failed.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		usr, err := users.GetOrCreate(ctx, user.User{
			Idents: []user.Ident{
				{Platform: user.PlatformEmail,
					Value: claims.Address,
				},
			},
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create session.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		sess, err := sessions.Create(ctx, usr.ID)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create session.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		apiClaims := auth.NewAPIClaims(sess.Owner, auth.AccountUser)
		access, err := secretKey.Sign(apiClaims)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to sign access token.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		auth.NewHTTPContext(w, r).SetSession(sess.Key)
		_ = json.NewEncoder(w).Encode(Token{
			Token:     access,
			ExpiresIn: int32(apiClaims.ExpiresIn().Seconds()),
		})
	}
}

func login(baseURL string, secretKey *auth.SecretKey, producer Producer, tubeEmail string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var loginRequest Login
		err := json.NewDecoder(r.Body).Decode(&loginRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := email.Validate(loginRequest.Address); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := secretKey.Sign(auth.NewEmailClaims(loginRequest.Address))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to create email token.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		msg, err := emails.Login(baseURL, loginRequest.Address, token)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to render login email.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		payload, err := proto.Marshal(&iam.SendEmail{
			Emails: []*iam.Email{{
				FromName:  msg.FromName,
				ToAddress: msg.ToAddress,
				Subject:   msg.Subject,
				Html:      msg.HTML,
			}},
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to marshal email.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := proto.Marshal(
			&queue.Job{
				Payload: payload,
				TraceId: trace.ID(ctx),
			},
		)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to marshal email.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := producer.Put(ctx, tubeEmail, body, beanstalk.PutParams{TTR: 10 * time.Second})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to put email job.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email login job emitted.")
	}
}

func logout(sessions *session.CRUD) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		authCtx := auth.NewHTTPContext(w, r)
		err := sessions.Delete(ctx, authCtx.Session())
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to delete session.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		authCtx.ResetSession()
		log.Ctx(ctx).Info().Str("sessionKey", authCtx.Session()).Msg("User logged out.")
	}
}
