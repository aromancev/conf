package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/iam"
	"github.com/aromancev/confa/internal/platform/api"
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
