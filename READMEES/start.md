# STARTUP PROJECT


## Backend

---

```shell

    tamagochi/: cd backend
    tamagochi/backend: export JWT_SECRET='this-is-some-jwt-secret-256'
    tamagochi/backend: go run ./cmd/api
```

this will start our backend Golang project

## Database

---
```shell
    tamagochi/: make env-up
```

! If database havent inititalazed yet !
```shell
    tamagochi/: make env-up
    tamagochi/: make migrate-create seq='Name of seq'
    tamagochi/: make migrate-action action='up/down'(fist start should be up)!!!
```
