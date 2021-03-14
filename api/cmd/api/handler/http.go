package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"github.com/prep/beanstalk"
	"github.com/pressly/goose"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/iam"
	"github.com/aromancev/confa/internal/platform/api"
	"github.com/aromancev/confa/internal/platform/email"
)

func (h *Handler) createConfa(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user, err := iam.Authenticate(r)
	if err != nil {
		_ = api.Unauthorised().Write(ctx, w)
		return
	}

	var request confa.Confa
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = api.BadRequest(api.CodeMalformedRequest, err.Error()).Write(ctx, w)
		return
	}

	conf, err := h.confaCRUD.Create(ctx, user.ID, request)
	switch {
	case errors.Is(err, confa.ErrValidation):
		_ = api.BadRequest(api.CodeInvalidRequest, err.Error()).Write(ctx, w)
		return
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create confa")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	_ = api.Created(conf).Write(ctx, w)
}
func (h *Handler) Confa(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user, err := iam.Authenticate(r)
	if err != nil {
		_ = api.Unauthorised().Write(ctx, w)
		return
	}
	confId, err := uuid.Parse(ps.ByName("confa_id"))
	if err != nil {
		_ = api.BadRequest(api.CodeInvalidRequest, err.Error()).Write(ctx, w)
		return
	}
	conf, err := h.confaCRUD.Fetch(ctx, confId, user.ID)
	switch {
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

func (h *Handler) reset(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	conn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		h.dbConf.Host,
		h.dbConf.Port,
		h.dbConf.User,
		h.dbConf.Password,
		h.dbConf.Database,
	)
	pg, err := sql.Open("postgres", conn)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to connect")
		_ = api.InternalError().Write(ctx, w)
		return
	}

	goose.SetVerbose(false)
	err = goose.SetDialect("postgres")
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to connect")
		_ = api.InternalError().Write(ctx, w)
		return
	}
	goose.SetTableName("goose_db_version")
	err = goose.Up(pg, "internal/migrations")
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to connect")
		_ = api.InternalError().Write(ctx, w)
		return
	}
}
