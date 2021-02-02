package room

import (
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/utils"
)

type RoomHandler struct {
	sessions types.SessionManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
}

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
) *RoomHandler {
	// Init

	return &RoomHandler{
		sessions: sessions,
		desktop:  desktop,
		capture:  capture,
	}
}

func (h *RoomHandler) Route(r chi.Router) {
	r.With(auth.AdminsOnly).Route("/broadcast", func(r chi.Router) {
		r.Get("/", h.broadcastStatus)
		r.Post("/start", h.boradcastStart)
		r.Post("/stop", h.boradcastStop)
	})

	r.With(auth.HostsOnly).Route("/clipboard", func(r chi.Router) {
		r.Get("/", h.clipboardGetText)
		r.Post("/", h.clipboardSetText)
		r.Get("/image.png", h.clipboardGetImage)

		// TODO: Refactor. xclip is failing to set propper target type
		// and this content is sent back to client as text in another
		// clipboard update. Therefore endpoint is not usable!
		//r.Post("/image", h.clipboardSetImage)

		// TODO: Refactor. If there would be implemented custom target
		// retrieval, this endpoint would be useful.
		//r.Get("/targets", h.clipboardGetTargets)
	})

	r.Route("/keyboard", func(r chi.Router) {
		r.Get("/map", h.keyboardMapGet)
		r.With(auth.HostsOnly).Post("/map", h.keyboardMapSet)

		r.Get("/modifiers", h.keyboardModifiersGet)
		r.With(auth.HostsOnly).Post("/modifiers", h.keyboardModifiersSet)
	})

	r.Route("/control", func(r chi.Router) {
		r.Get("/", h.controlStatus)
		r.Post("/request", h.controlRequest)
		r.Post("/release", h.controlRelease)

		r.With(auth.AdminsOnly).Post("/take", h.controlTake)
		r.With(auth.AdminsOnly).Post("/give", h.controlGive)
		r.With(auth.AdminsOnly).Post("/reset", h.controlReset)
	})

	r.Route("/screen", func(r chi.Router) {
		r.With(auth.CanWatchOnly).Get("/", h.screenConfiguration)
		r.With(auth.CanWatchOnly).Get("/shot.jpg", h.screenShotGet)
		r.With(auth.CanWatchOnly).Get("/cast.jpg", h.screenCastGet)

		r.With(auth.AdminsOnly).Post("/", h.screenConfigurationChange)
		r.With(auth.AdminsOnly).Get("/configurations", h.screenConfigurationsList)
	})

	r.With(h.uploadMiddleware).Route("/upload", func(r chi.Router) {
		r.Post("/drop", h.uploadDrop)
		r.Post("/dialog", h.uploadDialogPost)
		r.Delete("/dialog", h.uploadDialogClose)
	})
}

func (h *RoomHandler) uploadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := auth.GetSession(r)
		if !session.IsHost() && (!session.CanHost() || !h.sessions.ImplicitHosting()) {
			utils.HttpForbidden(w, "Without implicit hosting, only host can upload files.")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
