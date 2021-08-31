package controllers

import (
	"siot/api/middlewares"
)

func (s *Server) initializeRoutes() {

	// Home Route
	s.Router.HandleFunc("/api", middlewares.SetMiddlewareJSON(s.Home)).Methods("GET")

	// Login Route
	s.Router.HandleFunc("/api/login", middlewares.SetMiddlewareJSON(s.Login)).Methods("POST")

	// Confirmation user
	s.Router.HandleFunc("/api/users/confirmation", middlewares.SetMiddlewareJSON(s.ConfirmUser)).Methods("PUT")

	// Super admin routes
	// Users routes
	s.Router.HandleFunc("/api/users", middlewares.SetMiddlewareAuthentication(
		middlewares.SetMiddlewareIsSuperAdmin(s.DB, s.CreateUser))).Methods("POST")

	// s.Router.HandleFunc("/api/users", middlewares.SetMiddlewareAuthentication(s.GetUsers)).Methods("GET")
	// s.Router.HandleFunc("/api/users/{id}", middlewares.SetMiddlewareJSON(s.GetUser)).Methods("GET")
	// s.Router.HandleFunc("/api/users/{id}", middlewares.SetMiddlewareAuthentication(s.UpdateUser)).Methods("PUT")
	// s.Router.HandleFunc("/api/users/{id}", middlewares.SetMiddlewareAuthentication(s.DeleteUser)).Methods("DELETE")

	// Non admin users
	// Tenants routes
	s.Router.HandleFunc("/api/tenants", middlewares.SetMiddlewareAuthentication(s.CreateTenant)).Methods("POST")
	s.Router.HandleFunc("/api/tenants", middlewares.SetMiddlewareAuthentication(s.ListTenants)).Methods("GET")

	// Devices routes
	s.Router.HandleFunc("/api/{tenant_id}/devices",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(s.DB, s.CreateDevice))).Methods("POST")

	s.Router.HandleFunc("/api/{tenant_id}/devices",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(s.DB, s.ListDevices))).Methods("GET")

	s.Router.HandleFunc("/api/{tenant_id}/devices/{device_id}",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(
				s.DB, middlewares.SetMiddlewareIsDeviceValid(s.DB, s.ShowDevice)))).Methods("GET")

	s.Router.HandleFunc("/api/{tenant_id}/devices/{device_id}",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(
				s.DB, middlewares.SetMiddlewareIsDeviceValid(s.DB, s.UpdateDevice)))).Methods("PUT")

	s.Router.HandleFunc("/api/{tenant_id}/devices/{device_id}",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(
				s.DB, middlewares.SetMiddlewareIsDeviceValid(s.DB, s.DeleteDevice)))).Methods("DELETE")

	// Data routes
	s.Router.HandleFunc("/api/{tenant_id}/devices/{device_id}/data",
		middlewares.SetMiddlewareIsTenantValid(
			s.DB, middlewares.SetMiddlewareIsDeviceValidAndActive(s.DB, s.SendData))).Methods("POST")

	s.Router.HandleFunc("/api/{tenant_id}/devices/{device_id}/data", middlewares.SetMiddlewareAuthentication(
		middlewares.SetMiddlewareIsTenantValid(
			s.DB, middlewares.SetMiddlewareIsDeviceValid(s.DB, s.GetData)))).Methods("GET")
}
