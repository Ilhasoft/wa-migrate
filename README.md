# wa-migrate

## How run

1. Change the name of config.json.dev to config.json
3. Set the configs on config.json
2. Change the name of data.json.dev to data.json
    1. The slug will substitute the first %s of BaseURL
4. Set your data in the data.json
5. Run with `go run .`

## Examples

config.json
```json
{
    "BaseURL":    "https://my-%s.com",
	"Username":   "admin",
	"DataFile":   "./data.json",
	"BackupPath": "./backups"
}
```

data.json
```json
{
    "slug": "password",
    "my_slug": "password2",
}
```

It will touch the endpoints:
- https://my-slug.com
- https://my-my_slug.com