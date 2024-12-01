# Переменные окружения

> (*) - помечены обязательные для заполнения переменные окружения.

1. `APP_NAME` - название приложения, используется только при создании Docker контейнеров и никак не влияет на
   выполнение программы. Значение по умолчанию: `SiteFormChecker`;
2. `DB_CON` - драйвер для подключения к базе данных. Значение по умолчанию: `mysql`;
3. `DB_HOST` - хост для подключения к базе данных. Значение по умолчанию: `127.0.0.1`;
4. `DB_PORT` - порт для подключения к базе данных. Значение по умолчанию: `3306`;
5. `DB_USER` - пользователь для подключения к базе данных. Значение по умолчанию: `root`;
6. `DB_PASS` - пароль для подключения к базе данных. Значение по умолчанию: `отсутствует`;
7. `DB_NAME` - название подключаемой базы данных. Значение по умолчанию: `site_form_checker`;
8. `DEBUG_MODE` - режим отладки. При включенном режиме отладки в консоль выводиться дополнительная информация о
   проверках форм (Этап проверки формы, логи из консоли браузера и т.п.). Значение по умолчанию: `true`;
9. `MAX_GOROUTINES` - максимальное кол-во одновременно запущенных гоурутин для отправки форм. Значение по умолчанию:
   `10`;
10. `REMOTE_BROWSER_SCHEMA` - http схема для подключения к удаленному браузеру. Значение по умолчанию: `http`;
11. `REMOTE_BROWSER_URL` - адрес для подключения к удаленному браузеру. Значение по умолчанию: `127.0.0.1`;
12. `REMOTE_BROWSER_PORT` - порт для подключения к удаленному браузеру. Значение по умолчанию: `9222`;
13. `SEND_FORM_ATTEMPTS` - кол-во попыток для отправки формы. Значение по умолчанию: `3`;
14. `SEND_FORM_TIMEOUT` - общий тайм-аут для отправки формы (в секундах). Значение по умолчанию: `60`;
15. `SEND_FORM_RETRY_DELAY` - тайм-аут между попытками для отправки формы (в секундах). Значение по умолчанию: `10`;
16. `CRM_URL` - endpoint в CRM для проверки отправленной формы. Значение по умолчанию: `отсутствует`;
17. `CRM_TOKEN` - токен для Bearer Authorization. Значение по умолчанию: `отсутствует`;
18. `CRM_ATTEMPTS` - кол-во попыток для проверки отправленной формы в CRM. Значение по умолчанию: `3`;
19. `CRM_RETRY_DELAY` - тайм-аут между попытками для проверки отправленной формы в CRM (в секундах). Значение по
    умолчанию:
    `10`;
20. `TELEGRAM_CHAT_ID` - ID телеграм чата, в который будут приходить уведомления о проверках форм. Значение по
    умолчанию: `отсутствует`;
21. `TELEGRAM_TOKEN` - токен телеграм бота, который будет отправлять уведомления о проверках форм. Значение по
    умолчанию: `отсутствует`;
22. `TELEGRAM_PARSE_MODE` - режим отправки уведомлений о проверках форм в телеграм чат. Значение по умолчанию: `html`;

***

# База данных

* ### Таблица `forms`:

```sql
CREATE TABLE IF NOT EXISTS `forms`
(
    `id`                bigint(20)   NOT NULL,
    `name`              varchar(255) NOT NULL,
    `url`               varchar(255) NOT NULL,
    `element_for_click` varchar(255) NOT NULL,
    `expected_element`  varchar(255) NOT NULL,
    `submit_element`    varchar(255) NOT NULL,
    `result_element`    varchar(255) NOT NULL,
    `created_at`        timestamp    NULL DEFAULT current_timestamp(),
    `updated_at`        datetime          DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
```

* ### Таблица `inputs`:

```sql
CREATE TABLE IF NOT EXISTS `inputs`
(
    `id`         bigint(20)   NOT NULL,
    `form_id`    bigint(20)   NOT NULL,
    `selector`   varchar(255) NOT NULL,
    `value`      varchar(255) NOT NULL,
    `for_uuid`   tinyint(1)   NOT NULL DEFAULT 0,
    `created_at` timestamp    NULL     DEFAULT current_timestamp(),
    `updated_at` datetime              DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
```

***

# Модели

* ### Модель `Form`:
    * `Id` `Тип: Int` `Теги: 'db: "id" json: "id"'` - ID записи в базе данных;
    * `Name` `Тип: string` `Теги: 'db:"name" json:"name"'` - название формы;
    * `Url` `Тип: string` `Теги: 'db:"url" json:"url"'` - URL адрес, по которому осуществляется переход;
    * `ElemForClick` `Тип: string` `Теги: 'db:"element_for_click" json:"element_for_click"'` - JS селектор элемента, на
      который будет произведено нажатие;
    * `ExpElem` `Тип: string` `Теги: 'db:"expected_element" json:"expected_element"'` - JS селектор элемента, появление
      которого ожидается на странице;
    * `SubmitElem` `Тип: string` `Теги: 'db:"submit_element" json:"submit_element"'` - JS селектор элемента, по нажатию
      которого будет отправлена форма (относительно `ExpElem`);
    * `ResElem` `Тип: string` `Теги: 'db:"result_element" json:"result_element"'` - JS селектор элемента, появление
      которого ожидается на странице после отправки формы;
    * `CreatedAt` `Тип: string` `Теги: 'db:"created_at" json:"created_at"'` - дата создания записи в базе данных;
    * `UpdatedAt` `Тип: string` `Теги: 'db:"updated_at" json:"updated_at"'` - дата обновления записи в базе данных;

* ### Модель `Input`:
    * `Id` `Тип: Int` `Теги: 'db: "id" json: "id"'` - ID записи в базе данных;
    * `FormId` `Тип: Int` `Теги: 'db: "form_id" json: "form_id"'` - ID записи с формой, к которой привязан инпут, в базе
      данных;
    * `Selector` `Тип: string` `Теги: 'db:"selector" json:"selector"'` - JS селектор инпута для взаимодействия (
      относительно `ExpElem`);
    * `Value` `Тип: string` `Теги: 'db:"value" json:"value"'` - значение, которое будет указано в инпуте;
    * `ForUuid` `Тип: bool` `Теги: 'db:"for_uuid" json:"for_uuid"'` - необходимо ли указывать `uuid` в качестве значения
      инпута;
    * `CreatedAt` `Тип: string` `Теги: 'db:"created_at" json:"created_at"'` - дата создания записи в базе данных;
    * `UpdatedAt` `Тип: string` `Теги: 'db:"updated_at" json:"updated_at"'` - дата обновления записи в базе данных;

***

# Выполнение программы

***