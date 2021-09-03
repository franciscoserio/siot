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
		middlewares.SetMiddlewareIsSuperAdmin(s.DB, s.CreateAdminUser))).Methods("POST")

	// Admin user
	// Users routes
	s.Router.HandleFunc("/api/{tenant_id}/users", middlewares.SetMiddlewareAuthentication(
		middlewares.SetMiddlewareIsAdmin(
			s.DB, middlewares.SetMiddlewareIsTenantValid(s.DB, s.AddUser)))).Methods("POST")

	// Non admin users
	// Users routes
	s.Router.HandleFunc("/api/{tenant_id}/users",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(s.DB, s.GetTenantUsers))).Methods("GET")

	s.Router.HandleFunc("/api/{tenant_id}/users/{user_id}",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(
				s.DB, middlewares.SetMiddlewareIsUserTenantValid(s.DB, s.GetTenantUser)))).Methods("GET")

	// Tenants routes
	s.Router.HandleFunc("/api/tenants", middlewares.SetMiddlewareAuthentication(
		middlewares.SetMiddlewareIsAdmin(s.DB, s.CreateTenant))).Methods("POST")

	s.Router.HandleFunc("/api/tenants", middlewares.SetMiddlewareAuthentication(
		middlewares.SetMiddlewareIsAdmin(s.DB, s.ListTenants))).Methods("GET")

	s.Router.HandleFunc("/api/tenants/{tenant_id}",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsAdmin(
				s.DB, middlewares.SetMiddlewareIsTenantValid(s.DB, s.UpdateTenant)))).Methods("PUT")

	// Devices routes
	s.Router.HandleFunc("/api/{tenant_id}/devices",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(s.DB, s.CreateDevice))).Methods("POST")

	s.Router.HandleFunc("/api/tenants/{tenant_id}",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsAdmin(
				s.DB, middlewares.SetMiddlewareIsTenantValid(s.DB, s.GetTenant)))).Methods("GET")

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
		middlewares.SetMiddlewareIsDeviceValidAndActive(s.DB, s.SendData)).Methods("POST")

	s.Router.HandleFunc("/api/{tenant_id}/devices/{device_id}/data",
		middlewares.SetMiddlewareAuthentication(
			middlewares.SetMiddlewareIsTenantValid(
				s.DB, middlewares.SetMiddlewareIsDeviceValid(s.DB, s.GetData)))).Methods("GET")
}
