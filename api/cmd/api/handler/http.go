package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aromancev/confa/internal/user/ident"

	"github.com/aromancev/confa/internal/confa/talk"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/user/session"
)

func (h *Handler) createConfa(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	userID, err := auth.Authenticate(r)
	if err != nil {
		_ = api.Unauthorised().Write(ctx, w)
		return
	}

	var request confa.Confa
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = api.BadRequest(api.CodeMalformedRequest, err.Error()).Write(ctx, w)
		return
	}

	conf, err := h.confaCRUD.Create(ctx, userID, request)
	switch {
	case errors.Is(err, confa.ErrValidation):
		_ = api.BadRequest(api.CodeInvalidRequest, err.Error()).Write(ctx, w)
		return
	case errors.Is(err, confa.ErrDuplicatedEntry):
		_ = api.BadRequest(api.CodeDuplicatedEntry, err.Error()).Write(ctx, w)
		return
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create confa")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	_ = api.Created(conf).Write(ctx, w)
}

func (h *Handler) confa(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	confID, err := uuid.Parse(ps.ByName("confa_id"))
	if err != nil {
		_ = api.NotFound(err.Error()).Write(ctx, w)
		return
	}
	conf, err := h.confaCRUD.Fetch(ctx, confID)
	switch {
	case errors.Is(err, confa.ErrNotFound):
		_ = api.NotFound(err.Error()).Write(ctx, w)
		return
	case errors.Is(err, confa.ErrValidation):
		_ = api.BadRequest(api.CodeInvalidRequest, err.Error()).Write(ctx, w)
		return
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to fetch confa")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	_ = api.OK(conf).Write(ctx, w)
}

func (h *Handler) createTalk(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	userID, err := auth.Authenticate(r)
	if err != nil {
		_ = api.Unauthorised().Write(ctx, w)
		return
	}

	var request talk.Talk
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = api.BadRequest(api.CodeMalformedRequest, err.Error()).Write(ctx, w)
		return
	}

	confaID, err := uuid.Parse(ps.ByName("confa_id"))
	if err != nil {
		_ = api.NotFound(err.Error()).Write(ctx, w)
		return
	}

	tlk, err := h.talkCRUD.Create(ctx, confaID, userID, request)
	switch {
	case errors.Is(err, talk.ErrValidation):
		_ = api.BadRequest(api.CodeInvalidRequest, err.Error()).Write(ctx, w)
		return
	case errors.Is(err, talk.ErrDuplicatedEntry):
		_ = api.BadRequest(api.CodeDuplicatedEntry, err.Error()).Write(ctx, w)
		return
	case errors.Is(err, talk.ErrPermissionDenied):
		_ = api.Forbidden().Write(ctx, w)
		return
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create talk")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	_ = api.Created(tlk).Write(ctx, w)
}

func (h *Handler) talk(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	talkID, err := uuid.Parse(ps.ByName("talk_id"))
	if err != nil {
		_ = api.NotFound(err.Error()).Write(ctx, w)
		return
	}

	tlk, err := h.talkCRUD.Fetch(ctx, talkID)
	switch {
	case errors.Is(err, talk.ErrNotFound):
		_ = api.NotFound(err.Error()).Write(ctx, w)
		return
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to fetch talk")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	_ = api.OK(tlk).Write(ctx, w)
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	rawAuthentication := r.Header.Get("Authentication")
	authArray := strings.Split(rawAuthentication, " ")

	pltfrm, token := authArray[0], authArray[1]
	if pltfrm != "Email" {
		_ = api.Unauthorised("incorrect platform").Write(ctx, w)
		return
	}

	claims, err := h.verify.EmailToken(token)
	if err != nil {
		_ = api.Unauthorised().Write(ctx, w)
		return
	}

	idnt := ident.Ident{}
	idnt.Platform = ident.PlatformEmail
	idnt.Value = claims.Address

	userID, err := h.identCRUD.GetOrCreate(ctx, idnt)

	sess, err := h.sessionCRUD.Create(ctx, userID)
	switch {
	case errors.Is(err, session.ErrValidation):
		_ = api.BadRequest(api.CodeInvalidRequest, err.Error()).Write(ctx, w)
		return
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create session")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	_ = api.Created(sess).Write(ctx, w)
}

type loginReq struct {
	Email string `json:"email"`
}

func (r loginReq) Validate() error {
	if err := email.ValidateEmail(r.Email); err != nil {
		return err
	}
	return nil
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = api.BadRequest(api.CodeMalformedRequest, err.Error()).Write(ctx, w)
		return
	}

	if err := req.Validate(); err != nil {
		_ = api.BadRequest(api.CodeInvalidRequest, err.Error()).Write(ctx, w)
		return
	}

	token, err := h.sign.EmailToken(req.Email)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create email token")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	msg, err := emails.Login(h.baseURL, req.Email, token)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to render login email")
		_ = api.InternalError().Write(ctx, w)
		return
	}
	body, err := json.Marshal([]email.Email{msg})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal email")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	id, err := h.producer.Put(ctx, TubeEmail, body, beanstalk.PutParams{})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to put email job")
		_ = api.InternalError().Write(ctx, w)
		return
	}
	log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email login job emitted")
}
