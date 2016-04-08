package validator

import (
	"fmt"
	"net/http"

	"linksmart.eu/lc/sec/auth"
	"linksmart.eu/lc/sec/authz"
)

// HTTP Handler for service ticket validation
func (v *CASValidator) Handler(serverAddr, serviceID string, authz *authz.Conf, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		
		// No check for OPTIONS method
		if r.Method == "OPTIONS" {
			auth.Log.Printf("[%s] %q %s\n", r.Method, r.URL.String(), "")
			next.ServeHTTP(w, r)
			return
		}
		
		X_Auth_Token := r.Header.Get("X-Auth-Token")
		
		if X_Auth_Token == "" {
			auth.Log.Printf("[%s] %q %s\n", r.Method, r.URL.String(), "X-Auth-Token not specified.")
			auth.HTTPErrorResponse(http.StatusUnauthorized, "Unauthorized request: X-Auth-Token entity header not specified.", w)
			return
		}

		// Validate Token
		valid, body, err := v.Validate(serverAddr, serviceID, X_Auth_Token)
		if err != nil {
			auth.Log.Printf("[%s] %q %s\n", r.Method, r.URL.String(), "Authentication server error: "+err.Error())
			auth.HTTPErrorResponse(http.StatusInternalServerError, "Authentication server error: "+err.Error(), w)
			return
		}
		if !valid {
			if _, ok := body["error"]; ok {
				auth.Log.Printf("[%s] %q %s\n", r.Method, r.URL.String(), body["error"])
				auth.HTTPErrorResponse(http.StatusUnauthorized, "Unauthorized request: "+body["error"], w)
				return
			}
			auth.HTTPErrorResponse(http.StatusUnauthorized, "Unauthorized request.", w)
			return
		}

		// Check for optional authorization
		if authz != nil {
			// Check if user matches authorization rules
			authorized := authz.Authorized(r.URL.Path, r.Method, body["user"], body["group"])
			if !authorized {
				auth.Log.Printf("[%s] %q %s `%s`/`%s`\n", r.Method, r.URL.String(),
					"Access denied for", body["group"], body["user"])
				auth.HTTPErrorResponse(http.StatusForbidden,
					fmt.Sprintf("Access denied for `%s`/`%s`", body["group"], body["user"]), w)
				return
			}
		}

		// Valid token, proceed to next handler
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
