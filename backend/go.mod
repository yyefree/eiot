module eiot

go 1.24

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/spf13/viper v1.18.2
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.26.1
)

require (
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.16.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/twitchtv/twirp v5.9.0+incompatible // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.5.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gorm.io/driver/sqlite v1.6.0 // indirect
)

replace (
	gorm.io/driver/sqlite => modernc.org/sqlite v1.22.1
	gorm.io/gorm => gorm.io/gorm v1.26.1
)