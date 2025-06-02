## dice_roll - gRPC-сервис-игра "Подбрасывание кубика"

### Описание структуры проекта
<ul>
  <li><b>cmd/</b> - тут находится точка входы программы, + мигратор для миграций</li>  
  <li><b>config/</b> - директория с файлами конфига проекта</li>
  <li><b>gen/</b> - сгенерированные protoc файлы</li>
  <li><b>internal/</b> - внутренние "кишки" проекта</li>
  <li><b>internal/api</b> - реализация самого апи игры</li>
  <li><b>internal/app</b> - реализация создания gRPC сервера</li>
  <li><b>internal/config</b> - реализация чтения файла конфига</li>
  <li><b>internal/generators</b> - тут находятся роллеры, генераторы чисел для бросков</li>
  <li><b>internal/models</b> - модели объектов</li>
  <li><b>internal/storage</b> - реализация разных видов хранилищ</li>
  <li><b>internal/utils</b> - разное вспомогательное</li>
  <li><b>migrations/</b> - файлы миграций</li>
  <li><b>proto/</b> - директория с прото файлами</li>
  <li><b>tests/</b> - тесты</li>
</ul>

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

#### Описание логики работы

1. Клиент создает сессию. Сервер генерирует:
   1. `server_seed` (32 байта)
   2. `session_id` (UUID)
   3. возвращает `server_seed_hash = SHA256(server_seed)`
2. Клиент делает бросок. Сервер:
   1. Инкрементит nonce
   2. Выбирает случайный генератор
   3. Вычисляет HMAC-SHA256(server_seed + client_seed + nonce)
   4. Генерирует броски через выбранный генератор
   5. И в конце возвращает все параметры + раскрытый server_seed
3. Клиент верифицирует результат
