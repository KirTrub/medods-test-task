package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "github.com/KirTrub/medods-test-task/internal/domain/task"
	scheduleusecase "github.com/KirTrub/medods-test-task/internal/usecase/schedule"
)

type ScheduleHandler struct {
	usecase scheduleusecase.Usecase
}

func NewScheduleHandler(usecase scheduleusecase.Usecase) *ScheduleHandler {
	return &ScheduleHandler{usecase: usecase}
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := mutationDTOToCreateInput(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), input)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newScheduleDTO(created))
}

func (h *ScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getScheduleIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(s))
}

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getScheduleIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := mutationDTOToUpdateInput(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, input)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(updated))
}

func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getScheduleIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.usecase.List(r.Context())
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	response := make([]scheduleDTO, 0, len(schedules))
	for i := range schedules {
		response = append(response, newScheduleDTO(&schedules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getScheduleIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing schedule id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid schedule id")
	}

	return id, nil
}

func mutationDTOToCreateInput(req scheduleMutationDTO) (scheduleusecase.CreateInput, error) {
	dates, err := parseDates(req.Dates)
	if err != nil {
		return scheduleusecase.CreateInput{}, err
	}

	input := scheduleusecase.CreateInput{
		Title:        req.Title,
		Description:  req.Description,
		Type:         taskdomain.ScheduleType(req.Type),
		Dates:        dates,
		DeadlineDays: req.DeadlineDays,
	}

	if req.EveryNDays != nil {
		input.EveryNDays = *req.EveryNDays
	}
	if req.DayOfMonth != nil {
		input.DayOfMonth = *req.DayOfMonth
	}
	if req.Parity != nil {
		input.Parity = taskdomain.Parity(*req.Parity)
	}

	return input, nil
}

func mutationDTOToUpdateInput(req scheduleMutationDTO) (scheduleusecase.UpdateInput, error) {
	dates, err := parseDates(req.Dates)
	if err != nil {
		return scheduleusecase.UpdateInput{}, err
	}

	input := scheduleusecase.UpdateInput{
		Title:        req.Title,
		Description:  req.Description,
		Type:         taskdomain.ScheduleType(req.Type),
		Dates:        dates,
		DeadlineDays: req.DeadlineDays,
	}

	if req.EveryNDays != nil {
		input.EveryNDays = *req.EveryNDays
	}
	if req.DayOfMonth != nil {
		input.DayOfMonth = *req.DayOfMonth
	}
	if req.Parity != nil {
		input.Parity = taskdomain.Parity(*req.Parity)
	}

	return input, nil
}

func parseDates(strs []string) ([]time.Time, error) {
	if len(strs) == 0 {
		return nil, nil
	}
	result := make([]time.Time, 0, len(strs))
	for _, s := range strs {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, fmt.Errorf("invalid date %q: expected format YYYY-MM-DD", s)
		}
		result = append(result, t.UTC())
	}
	return result, nil
}

func writeScheduleUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrScheduleNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, scheduleusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
