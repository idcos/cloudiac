module cloudiac

go 1.16

replace github.com/google/flatbuffers v1.12.0 => github.com/google/flatbuffers v1.12.1

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Shopify/sarama v1.28.0
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/casbin/casbin/v2 v2.31.9
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/elazarl/goproxy v0.0.0-20210801061803-8e322dfb79c4 // indirect
	github.com/gin-contrib/sse v0.1.0
	github.com/gin-gonic/gin v1.7.2
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/google/go-github v17.0.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/consul/api v1.8.1
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/joho/godotenv v1.3.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.4 // indirect
	github.com/lib/pq v1.10.2
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/open-policy-agent/opa v0.32.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/rs/xid v1.2.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.7.0
	github.com/unliar/utils v0.1.1
	github.com/xanzy/go-gitlab v0.47.0
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.1.1
	gorm.io/gorm v1.21.12
	gorm.io/plugin/soft_delete v1.0.2
	moul.io/http2curl v1.0.0 // indirect
)
