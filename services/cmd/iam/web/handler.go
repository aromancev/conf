package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/google/gsi"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/confa"
	"github.com/aromancev/confa/internal/proto/queue"
	"github.com/aromancev/confa/internal/proto/sender"
	"github.com/aromancev/confa/internal/routes"
	"github.com/aromancev/confa/session"
	"github.com/aromancev/confa/user"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Token struct {
	Token     string `json:"token"`
	ExpiresIn int32  `json:"expiresIn"`
}

type CreateSessionByEmail struct {
	EmailToken string `json:"emailToken"`
}

type CreateSessionByCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateSessionByGoogleID struct {
	Token string `json:"token"`
}

type UpdatePassword struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	Logout      bool   `json:"logout"`
}

type ResetPassword struct {
	EmailToken string `json:"emailToken"`
	Password   string `json:"password"`
	Logout     bool   `json:"logout"`
}

type CreatePassword struct {
	EmailToken string `json:"emailToken"`
	Password   string `json:"password"`
}

type Email struct {
	Address string `json:"address"`
}

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Handler struct {
	tubes       Tubes
	auth        *Auth
	router      *http.ServeMux
	publicKey   *auth.PublicKey
	secretKey   *auth.SecretKey
	pages       *routes.Pages
	sessions    *session.Mongo
	user        *user.Actions
	producer    Producer
	googlePK    *gsi.PublicKey
	googleCreds gsi.Creds
}

type Tubes struct {
	Send         string
	UpdateAvatar string
}

func NewHandler(
	tubes Tubes,
	httpAuth *Auth,
	pages *routes.Pages,
	secretKey *auth.SecretKey,
	publicKey *auth.PublicKey,
	resolver *Resolver,
	sessions *session.Mongo,
	userActions *user.Actions,
	producer Producer,
	googlePK *gsi.PublicKey,
	googleCreds gsi.Creds,
) *Handler {
	h := &Handler{
		tubes:       tubes,
		auth:        httpAuth,
		router:      http.NewServeMux(),
		publicKey:   publicKey,
		secretKey:   secretKey,
		pages:       pages,
		sessions:    sessions,
		user:        userActions,
		producer:    producer,
		googlePK:    googlePK,
		googleCreds: googleCreds,
	}

	// All routes must be on the first level in order for secure cookies to work.
	h.router.HandleFunc("/health", ok)
	h.router.Handle(
		"/graph",
		withHTTPAuth(
			&relay.Handler{
				Schema: graphql.MustParseSchema(schema, resolver, graphql.UseFieldResolvers()),
			},
		),
	)
	h.router.HandleFunc(
		"/token",
		h.fetchToken,
	)
	h.router.HandleFunc(
		"/session-email",
		h.createSessionByEmail,
	)
	h.router.HandleFunc(
		"/session-credentials",
		h.createSessionByCredentials,
	)
	h.router.HandleFunc(
		"/session-google-sign-in",
		h.createSessionByGoogleSignIn,
	)
	h.router.HandleFunc(
		"/email-login",
		h.emailLogin,
	)
	h.router.HandleFunc(
		"/email-create-password",
		h.emailCreatePassword,
	)
	h.router.HandleFunc(
		"/email-reset-password",
		h.emailResetPassword,
	)
	h.router.HandleFunc(
		"/logout",
		h.logout,
	)
	h.router.HandleFunc(
		"/password-create",
		h.createPassword,
	)
	h.router.HandleFunc(
		"/password-update",
		h.updatePassword,
	)
	h.router.HandleFunc(
		"/password-reset",
		h.resetPassword,
	)
	return h
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
	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func ok(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func (h *Handler) fetchToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, err := h.sessions.FetchOne(ctx, session.Lookup{Key: h.auth.Session(r)})
	if err != nil {
		if !errors.Is(err, session.ErrNotFound) && !errors.Is(err, session.ErrValidation) {
			log.Ctx(ctx).Err(err).Msg("Failed to fetch session.")
		}
		var claims auth.APIClaims
		err := h.publicKey.Verify(h.auth.GuestToken(r), &claims)
		if err != nil {
			// Session not found, try validating guest claims.
			guestClaims := auth.NewAPIClaims(uuid.New(), auth.AccountGuest)
			guestToken, err := h.secretKey.Sign(guestClaims)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to sign API token.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			h.auth.SetGuestToken(w, guestToken)
			_ = json.NewEncoder(w).Encode(Token{
				Token:     guestToken,
				ExpiresIn: int32(claims.ExpiresIn().Seconds()),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(Token{
			Token:     h.auth.GuestToken(r),
			ExpiresIn: int32(claims.ExpiresIn().Seconds()),
		})
		return
	}

	userClaims := auth.NewAPIClaims(sess.Owner, auth.AccountUser)
	userToken, err := h.secretKey.Sign(userClaims)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign API token.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(Token{
		Token:     userToken,
		ExpiresIn: int32(userClaims.ExpiresIn().Seconds()),
	})
}

func (h *Handler) createSessionByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request CreateSessionByEmail
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var claims auth.EmailClaims
	err = h.publicKey.Verify(request.EmailToken, &claims)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := h.user.GetOrCreate(ctx, user.User{
		Idents: []user.Ident{
			{
				Platform: user.PlatformEmail,
				Value:    claims.Address,
			},
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to find user.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	created, err := h.sessions.Create(ctx, session.Session{
		Key:   session.NewKey(),
		Owner: usr.ID,
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create session.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sess := created[0]

	apiClaims := auth.NewAPIClaims(sess.Owner, auth.AccountUser)
	access, err := h.secretKey.Sign(apiClaims)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign access token.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := Token{
		Token:     access,
		ExpiresIn: int32(apiClaims.ExpiresIn().Seconds()),
	}
	h.auth.ResetGuestToken(w)
	h.auth.SetSession(w, sess.Key)
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handler) createSessionByCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request CreateSessionByCredentials
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	usr, err := h.user.CheckPassword(
		ctx,
		user.Ident{
			Platform: user.PlatformEmail,
			Value:    request.Email,
		},
		user.Password(request.Password),
	)
	switch {
	case errors.Is(err, user.ErrNotFound), errors.Is(err, user.ErrValidation):
		w.WriteHeader(http.StatusNotFound)
		return
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to check user password.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	created, err := h.sessions.Create(ctx, session.Session{
		Key:   session.NewKey(),
		Owner: usr.ID,
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create session.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sess := created[0]

	apiClaims := auth.NewAPIClaims(sess.Owner, auth.AccountUser)
	access, err := h.secretKey.Sign(apiClaims)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign access token.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.auth.ResetGuestToken(w)
	h.auth.SetSession(w, sess.Key)
	_ = json.NewEncoder(w).Encode(Token{
		Token:     access,
		ExpiresIn: int32(apiClaims.ExpiresIn().Seconds()),
	})
}

func (h *Handler) createSessionByGoogleSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request CreateSessionByGoogleID
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pk, err := h.googlePK.PEM(ctx, gsi.KeyID(request.Token))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to get PEM.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	id, err := gsi.NewIDAuth(pk, h.googleCreds.ClientID).Verify(request.Token)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to validate token.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := h.user.GetOrCreate(ctx, user.User{
		Idents: []user.Ident{
			{
				Platform: user.PlatformEmail,
				Value:    id.Email,
			},
		},
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to find user.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	created, err := h.sessions.Create(ctx, session.Session{
		Key:   session.NewKey(),
		Owner: usr.ID,
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create session.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sess := created[0]

	apiClaims := auth.NewAPIClaims(sess.Owner, auth.AccountUser)
	access, err := h.secretKey.Sign(apiClaims)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign access token.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: Remove this when migrated to event-based message queue and create a new event.
	err = h.emitUpdateProfile(ctx, usr.ID, id.GivenName, id.FamilyName, id.Picture)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit profile update event.")
	}

	h.auth.ResetGuestToken(w)
	h.auth.SetSession(w, sess.Key)
	_ = json.NewEncoder(w).Encode(Token{
		Token:     access,
		ExpiresIn: int32(apiClaims.ExpiresIn().Seconds()),
	})
}

func (h *Handler) emailLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request Email
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := email.Validate(request.Address); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.secretKey.Sign(auth.NewEmailClaims(request.Address))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create email token.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := proto.Marshal(&sender.Send{
		Message: &sender.Message{
			Message: &sender.Message_Login_{
				Login: &sender.Message_Login{
					SecretUrl: h.pages.Login(routes.ActionLogin, token),
				},
			},
		},
		Delivery: &sender.Delivery{
			Delivery: &sender.Delivery_Email_{
				Email: &sender.Delivery_Email{
					ToAddress: request.Address,
				},
			},
		},
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

	id, err := h.producer.Put(ctx, h.tubes.Send, body, beanstalk.PutParams{TTR: time.Minute})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to put email job.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email login job emitted.")
}

func (h *Handler) emailCreatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request Email
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := email.Validate(request.Address); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.secretKey.Sign(auth.NewEmailClaims(request.Address))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create email token.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := proto.Marshal(&sender.Send{
		Message: &sender.Message{
			Message: &sender.Message_CreatePassword_{
				CreatePassword: &sender.Message_CreatePassword{
					SecretUrl: h.pages.Login(routes.ActionCreatePassword, token),
				},
			},
		},
		Delivery: &sender.Delivery{
			Delivery: &sender.Delivery_Email_{
				Email: &sender.Delivery_Email{
					ToAddress: request.Address,
				},
			},
		},
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

	id, err := h.producer.Put(ctx, h.tubes.Send, body, beanstalk.PutParams{TTR: time.Minute})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to put email job.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email create password job emitted.")
}

func (h *Handler) emailResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request Email
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := email.Validate(request.Address); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.secretKey.Sign(auth.NewEmailClaims(request.Address))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create email token.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := proto.Marshal(&sender.Send{
		Message: &sender.Message{
			Message: &sender.Message_ResetPassword_{
				ResetPassword: &sender.Message_ResetPassword{
					SecretUrl: h.pages.Login(routes.ActionResetPassword, token),
				},
			},
		},
		Delivery: &sender.Delivery{
			Delivery: &sender.Delivery_Email_{
				Email: &sender.Delivery_Email{
					ToAddress: request.Address,
				},
			},
		},
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

	id, err := h.producer.Put(ctx, h.tubes.Send, body, beanstalk.PutParams{TTR: time.Minute})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to put email job.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email reset password job emitted.")
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, err := h.sessions.Delete(ctx, session.Lookup{Key: h.auth.Session(r)})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to delete session.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.auth.ResetGuestToken(w)
	h.auth.ResetSession(w)
	log.Ctx(ctx).Info().Str("sessionKey", h.auth.Session(r)).Msg("User logged out.")
}

func (h *Handler) createPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request CreatePassword
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to unmarshal session.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var claims auth.EmailClaims
	err = h.publicKey.Verify(request.EmailToken, &claims)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Email verification failed.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = h.user.CreatePassword(
		ctx,
		user.Ident{
			Platform: user.PlatformEmail,
			Value:    claims.Address,
		},
		user.Password(request.Password),
	)
	switch {
	case errors.Is(err, user.ErrValidation):
		w.WriteHeader(http.StatusBadRequest)
		return
	case errors.Is(err, user.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Ctx(ctx).Info().Str("sessionKey", h.auth.Session(r)).Msg("Password created.")
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var claims auth.APIClaims
	if err := h.publicKey.Verify(h.auth.Token(r), &claims); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request UpdatePassword
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.user.UpdatePassword(
		ctx,
		claims.UserID,
		user.Password(request.OldPassword),
		user.Password(request.NewPassword),
	)
	switch {
	case errors.Is(err, user.ErrValidation):
		w.WriteHeader(http.StatusBadRequest)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if request.Logout {
		_, err = h.sessions.Delete(ctx, session.Lookup{Owner: claims.UserID})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to delete session.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.auth.ResetSession(w)
		log.Ctx(ctx).Info().Str("sessionKey", h.auth.Session(r)).Msg("Deleted all sessions.")
	}

	log.Ctx(ctx).Info().Str("sessionKey", h.auth.Session(r)).Msg("Password reset.")
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request ResetPassword
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to unmarshal session.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var claims auth.EmailClaims
	err = h.publicKey.Verify(request.EmailToken, &claims)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Email verification failed.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := h.user.ResetPassword(
		ctx,
		user.Ident{
			Platform: user.PlatformEmail,
			Value:    claims.Address,
		},
		user.Password(request.Password),
	)
	switch {
	case errors.Is(err, user.ErrValidation):
		w.WriteHeader(http.StatusBadRequest)
		return
	case errors.Is(err, user.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if request.Logout {
		_, err = h.sessions.Delete(ctx, session.Lookup{Owner: usr.ID})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to delete session.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.auth.ResetSession(w)
		log.Ctx(ctx).Info().Str("sessionKey", h.auth.Session(r)).Msg("Deleted all sessions.")
	}

	log.Ctx(ctx).Info().Str("sessionKey", h.auth.Session(r)).Msg("Password reset.")
}

// TODO: Remove this when migrated to event-based message queue and create a new event.
func (h *Handler) emitUpdateProfile(ctx context.Context, userID uuid.UUID, givenName, familyName, thumbnailURL string) error {
	id, _ := userID.MarshalBinary()
	job := confa.UpdateProfile{
		UserId:       id,
		GivenName:    givenName,
		FamilyName:   familyName,
		SkipIfExists: true,
	}
	if thumbnailURL != "" {
		job.Thumbnail = &confa.UpdateProfile_FileSource{
			PublicUrl: &confa.UpdateProfile_FileSource_PublicURL{
				Url: thumbnailURL,
			},
		}
	}
	payload, err := proto.Marshal(&job)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	body, err := proto.Marshal(
		&queue.Job{
			Payload: payload,
			TraceId: trace.ID(ctx),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	_, err = h.producer.Put(ctx, h.tubes.UpdateAvatar, body, beanstalk.PutParams{
		Delay: 2 * time.Second,
		TTR:   2 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}
	return nil
}
