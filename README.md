## dice_roll - gRPC-сервис-игра "Подбрасывание кубика"

#### Установка
1. Скопировать репозиторий, перейти в папку
```
git clone git@github.com:artemmj/dice_roll_v2.git
cd dice_roll_v2/
```
2. Создать файл .env с переменными окружения для докера (есть пример .env_example). При первом запуске создастся новая БД в контейнере, применятся миграции, пробросится volume. В .env_example указаны дефолтные параметры подключения, при необходимости можно (нужно) поменять.
```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres
```

3. Выполнить команду ```docker-compose up -d --build``` чтобы собрать контейнеры
  
4. Чтобы сыграть, используем например Postman, подключимся к localhost:50051 by gRPC и создадим сессию, сделав запрос на CreateSession, передав в теле запроса client_seed:
<img width="886" alt="Снимок экрана 2025-06-02 в 15 22 50" src="https://github.com/user-attachments/assets/521becbe-97f6-446a-8e4d-d0f8a5c665d9" />
В ответе вернется session_id и server_seed_hash

5. Далее начинаем игру, берем session_id и с ним идем на ручку Play, передав в теле запроса:
<img width="891" alt="Снимок экрана 2025-06-02 в 15 24 19" src="https://github.com/user-attachments/assets/42abb457-0f61-4826-9abf-74e7322373c4" />

в ответе вернется инфо об игре

6. Для валидации нужно взять клиентский сид, ожидаемое значение, имя генератора, номер раунда и так же раскрытый серверный сид (все это есть в ответе от Play), и пойти с этим на ручку VerifyRoll. В ответ вернется буль - успешна ли валидация или нет:
<img width="878" alt="Снимок экрана 2025-06-02 в 15 27 38" src="https://github.com/user-attachments/assets/29d76ffa-cb5b-4212-98bd-b1c02a44f902" />

7. Можно подключиться к контейнеру БД командой
```docker-compose exec db psql -U postgres```
Далее подключиться к БД postgres
```\c postgres;```
и убедиться, что в таблице с результатами game_results есть данные
```select * from game_results;```
<img width="717" alt="Снимок экрана 2025-06-02 в 15 28 35" src="https://github.com/user-attachments/assets/ba558165-1c24-4af1-8186-43f259e59f33" />

8. Чтобы пройти тесты, выполните команду при работающем контейнере
```
go test ./tests -count=1 -v
```
