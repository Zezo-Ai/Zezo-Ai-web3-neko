package room

import (
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
	"demodesk/neko/internal/utils"
	"demodesk/neko/internal/http/auth"
)

type ControlStatusPayload struct {
	HasHost bool `json:"has_host"`
	HostId string `json:"host_id,omitempty"`
}

type ControlTargetPayload struct {
	ID string `json:"id"`
}

func (h *RoomHandler) controlStatus(w http.ResponseWriter, r *http.Request) {
	host := h.sessions.GetHost()

	if host == nil {
		utils.HttpSuccess(w, ControlStatusPayload{
			HasHost: false,
		})
	} else {
		utils.HttpSuccess(w, ControlStatusPayload{
			HasHost: true,
			HostId: host.ID(),
		})
	}
}

func (h *RoomHandler) controlRequest(w http.ResponseWriter, r *http.Request) {
	host := h.sessions.GetHost()
	if host != nil {
		utils.HttpUnprocessableEntity(w, "There is already a host.")
		return
	}

	session := auth.GetSession(r)
	if !session.CanHost() {
		utils.HttpBadRequest(w, "Member is not allowed to host.")
		return
	}

	h.sessions.SetHost(session)

	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: true,
			HostID:  session.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlRelease(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.IsHost() {
		utils.HttpUnprocessableEntity(w, "Member is not the host.")
		return
	}

	if !session.CanHost() {
		utils.HttpBadRequest(w, "Member is not allowed to host.")
		return
	}

	h.desktop.ResetKeys()
	h.sessions.ClearHost()

	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: false,
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlTake(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r)
	if !session.CanHost() {
		utils.HttpBadRequest(w, "Member is not allowed to host.")
		return
	}

	h.sessions.SetHost(session)

	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: true,
			HostID:  session.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlGive(w http.ResponseWriter, r *http.Request) {
	memberId := chi.URLParam(r, "memberId")

	target, ok := h.sessions.Get(memberId)
	if !ok {
		utils.HttpNotFound(w, "Target member was not found.")
		return
	}

	if !target.CanHost() {
		utils.HttpBadRequest(w, "Target member is not allowed to host.")
		return
	}

	h.sessions.SetHost(target)

	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: true,
			HostID:  target.ID(),
		}, nil)

	utils.HttpSuccess(w)
}

func (h *RoomHandler) controlReset(w http.ResponseWriter, r *http.Request) {
	host := h.sessions.GetHost()
	if host == nil {
		utils.HttpSuccess(w)
		return
	}

	h.desktop.ResetKeys()
	h.sessions.ClearHost()

	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: false,
		}, nil)

	utils.HttpSuccess(w)
}
