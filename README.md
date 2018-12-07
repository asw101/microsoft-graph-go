# microsoft-graph-go

```bash
export AZ_CLIENT_ID='...'
export AZ_CLIENT_SECRET='...'
export AZ_TENANT_ID='...'
export AZ_GRAPH_SCOPES='User.Read' # or 'openid+profile'

# cli
go run main.go

# web
MODE=web go run main.go

# browser
http://localhost:8080/login
# redirects to: http://localhost:8080/auth

http://localhost:8080/token

http://localhost:8080/me


```