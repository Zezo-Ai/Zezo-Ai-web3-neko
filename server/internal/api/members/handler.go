package members

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type key int

const keyMemberCtx key = iota

type MembersHandler struct {
	members types.MemberManager
}

func New(
	members types.MemberManager,
) *MembersHandler {
	// Init

	return &MembersHandler{
		members: members,
	}
}

func (h *MembersHandler) Route(r types.Router) {
	r.Get("/", h.membersList)

	r.With(auth.AdminsOnly).Group(func(r types.Router) {
		r.Post("/", h.membersCreate)
		r.With(h.ExtractMember).Route("/{memberId}", func(r types.Router) {
			r.Get("/", h.membersRead)
			r.Post("/", h.membersUpdateProfile)
			r.Post("/password", h.membersUpdatePassword)
			r.Delete("/", h.membersDelete)
		})
	})
}

func (h *MembersHandler) RouteBulk(r types.Router) {
	r.With(auth.AdminsOnly).Group(func(r types.Router) {
		r.Post("/update", h.membersBulkUpdate)
		r.Post("/delete", h.membersBulkDelete)
	})
}

type MemberData struct {
	ID      string
	Profile types.MemberProfile
}

func SetMember(r *http.Request, session MemberData) context.Context {
	return context.WithValue(r.Context(), keyMemberCtx, session)
}

func GetMember(r *http.Request) MemberData {
	return r.Context().Value(keyMemberCtx).(MemberData)
}

func (h *MembersHandler) ExtractMember(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	memberId := chi.URLParam(r, "memberId")

	profile, err := h.members.Select(memberId)
	if err != nil {
		if errors.Is(err, types.ErrMemberDoesNotExist) {
			return nil, utils.HttpNotFound("member not found")
		}

		return nil, utils.HttpInternalServerError().WithInternalErr(err)
	}

	return SetMember(r, MemberData{
		ID:      memberId,
		Profile: profile,
	}), nil
}
