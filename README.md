Запускается либо через cmd/server/main.go (установив все зависимости через go mod), либо через docker-compose build. Перед запуском стоит посмотреть конфигурацию.

Три эндпоинта - 

1. Регистрация пользователя (POST /register):

	curl -X POST http://localhost:8080/register \
	     -H "Content-Type: application/json" \
	     -d '{"email": "user@example.com"}'

Ожидаемый ответ:

	{
	  "user_id": "0b4d5429-0251-4816-b90a-d481eb1ef633",
	  "access": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
	  "refresh": "H7VCtkp9NEE-SOVscj-FTbe97Ks5k9-dYIXQCUlM9Uk"
	}

2. Получение токенов (GET /login):

	curl -X GET "http://localhost:8080/login?user_id=0b4d5429-0251-4816-b90a-d481eb1ef633"

Ожидаемый ответ:

	{
	  "access": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
	  "refresh": "Pvza5ZqdDQY_-0N-aFfQhy3GWQUlv2Goe5AQuNyR75E"
	}


3. Обновление токенов (PATCH /refresh):

	curl -X PATCH http://localhost:8080/refresh \
	     -H "Content-Type: application/json" \
	     -d '{
	           "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
	           "refresh_token": "Pvza5ZqdDQY_-0N-aFfQhy3GWQUlv2Goe5AQuNyR75E"
	         }'

Ожидаемый ответ:

	{
	  "access": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
	  "refresh": "PAIjfTBsjKAc2dsrbNMm2U-ABZynAM_icvgxTW4wWds"
	}
