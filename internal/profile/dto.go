package profile

// CreateRequest represents the request to create a new profile
type CreateRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=64"`
	Algorithm string `json:"algorithm" binding:"required"`
	Limit     int    `json:"limit" binding:"required,min=1"`
	Window    string `json:"window" binding:"required"`
	Capacity  *int   `json:"capacity,omitempty"`
}

// UpdateRequest represents the request to update an existing profile
type UpdateRequest struct {
	Limit    *int    `json:"limit,omitempty"`
	Window   *string `json:"window,omitempty"`
	Capacity *int    `json:"capacity,omitempty"`
}

// Response represents a profile in API responses
type Response struct {
	Name      string `json:"name"`
	Algorithm string `json:"algorithm"`
	Limit     int    `json:"limit"`
	Window    string `json:"window"`
	Capacity  int    `json:"capacity"`
}

// ListResponse represents the response for listing profiles
type ListResponse struct {
	Profiles []Response `json:"profiles"`
}

// ToResponse converts a Profile entity to a Response DTO
func ToResponse(p *Profile) Response {
	return Response{
		Name:      p.Name,
		Algorithm: p.Algorithm,
		Limit:     p.Limit,
		Window:    p.Window,
		Capacity:  p.GetCapacity(),
	}
}

// ToListResponse converts a slice of Profile entities to a ListResponse
func ToListResponse(profiles []Profile) ListResponse {
	resp := ListResponse{
		Profiles: make([]Response, 0, len(profiles)),
	}
	for i := range profiles {
		resp.Profiles = append(resp.Profiles, ToResponse(&profiles[i]))
	}
	return resp
}
