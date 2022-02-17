package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/internal/user/session"
	"github.com/aromancev/confa/proto/queue"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var claims auth.EmailClaims
		err = publicKey.Verify(sessionRequest.EmailToken, &claims)
		if err != nil {
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

func login(baseURL string, secretKey *auth.SecretKey, producer Producer) http.HandlerFunc {
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

		body, err := queue.Marshal(&queue.EmailJob{
			Emails: []*queue.Email{{
				FromName:  msg.FromName,
				ToAddress: msg.ToAddress,
				Subject:   msg.Subject,
				Html:      msg.HTML,
			}},
		}, trace.ID(ctx))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to marshal email.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := producer.Put(ctx, queue.TubeEmail, body, beanstalk.PutParams{TTR: 10 * time.Second})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to put email job.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email login job emitted.")
	}
}
