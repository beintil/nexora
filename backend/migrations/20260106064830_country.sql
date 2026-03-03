-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS country (
    id SERIAL PRIMARY KEY,

    -- code — ISO 3166-1 alpha-2 код страны
    -- Используется всеми телефониями (Twilio, SIP, GSM и т.д.)
    -- Примеры: US, TH, RU
    code CHAR(2) NOT NULL UNIQUE,

    -- name — название страны на английском
    -- Используется как canonical name
    -- Примеры: United States of America, Thailand
    name VARCHAR(255) NOT NULL,

    -- description — локализованное название (например, на русском)
    -- Используется для UI / отчетов
    description VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Комментарии
COMMENT ON TABLE country IS 'Справочник стран мира (ISO 3166-1 alpha-2), используется всеми телефониями';

COMMENT ON COLUMN country.id IS 'Внутренний идентификатор страны';
COMMENT ON COLUMN country.code IS 'ISO 3166-1 alpha-2 код страны (US, TH, RU)';
COMMENT ON COLUMN country.name IS 'Название страны на английском';
COMMENT ON COLUMN country.description IS 'Локализованное название страны (например, на русском)';
COMMENT ON COLUMN country.created_at IS 'Дата создания записи';
COMMENT ON COLUMN country.updated_at IS 'Дата последнего обновления записи';

-- Индексы
CREATE UNIQUE INDEX IF NOT EXISTS idx_country_code ON country(code);
CREATE INDEX IF NOT EXISTS idx_country_name ON country(name);

INSERT INTO country (code, name, description) VALUES
    ('AF', 'Afghanistan', 'Афганистан'),
    ('AL', 'Albania', 'Албания'),
    ('DZ', 'Algeria', 'Алжир'),
    ('AS', 'American Samoa', 'Американское Самоа'),
    ('AD', 'Andorra', 'Андорра'),
    ('AO', 'Angola', 'Ангола'),
    ('AI', 'Anguilla', 'Ангилья'),
    ('AQ', 'Antarctica', 'Антарктида'),
    ('AG', 'Antigua and Barbuda', 'Антигуа и Барбуда'),
    ('AR', 'Argentina', 'Аргентина'),
    ('AM', 'Armenia', 'Армения'),
    ('AW', 'Aruba', 'Аруба'),
    ('AU', 'Australia', 'Австралия'),
    ('AT', 'Austria', 'Австрия'),
    ('AZ', 'Azerbaijan', 'Азербайджан'),

    ('BS', 'Bahamas', 'Багамы'),
    ('BH', 'Bahrain', 'Бахрейн'),
    ('BD', 'Bangladesh', 'Бангладеш'),
    ('BB', 'Barbados', 'Барбадос'),
    ('BY', 'Belarus', 'Беларусь'),
    ('BE', 'Belgium', 'Бельгия'),
    ('BZ', 'Belize', 'Белиз'),
    ('BJ', 'Benin', 'Бенин'),
    ('BM', 'Bermuda', 'Бермуды'),
    ('BT', 'Bhutan', 'Бутан'),
    ('BO', 'Bolivia', 'Боливия'),
    ('BA', 'Bosnia and Herzegovina', 'Босния и Герцеговина'),
    ('BW', 'Botswana', 'Ботсвана'),
    ('BR', 'Brazil', 'Бразилия'),
    ('IO', 'British Indian Ocean Territory', 'Британская территория в Индийском океане'),
    ('BN', 'Brunei', 'Бруней'),
    ('BG', 'Bulgaria', 'Болгария'),
    ('BF', 'Burkina Faso', 'Буркина-Фасо'),
    ('BI', 'Burundi', 'Бурунди'),

    ('KH', 'Cambodia', 'Камбоджа'),
    ('CM', 'Cameroon', 'Камерун'),
    ('CA', 'Canada', 'Канада'),
    ('CV', 'Cape Verde', 'Кабо-Верде'),
    ('KY', 'Cayman Islands', 'Каймановы острова'),
    ('CF', 'Central African Republic', 'Центральноафриканская Республика'),
    ('TD', 'Chad', 'Чад'),
    ('CL', 'Chile', 'Чили'),
    ('CN', 'China', 'Китай'),
    ('CO', 'Colombia', 'Колумбия'),
    ('KM', 'Comoros', 'Коморы'),
    ('CG', 'Congo', 'Конго'),
    ('CD', 'Congo (Democratic Republic)', 'Демократическая Республика Конго'),
    ('CR', 'Costa Rica', 'Коста-Рика'),
    ('CI', 'Côte d’Ivoire', 'Кот-д’Ивуар'),
    ('HR', 'Croatia', 'Хорватия'),
    ('CU', 'Cuba', 'Куба'),
    ('CY', 'Cyprus', 'Кипр'),
    ('CZ', 'Czech Republic', 'Чехия'),

    ('DK', 'Denmark', 'Дания'),
    ('DJ', 'Djibouti', 'Джибути'),
    ('DM', 'Dominica', 'Доминика'),
    ('DO', 'Dominican Republic', 'Доминиканская Республика'),

    ('EC', 'Ecuador', 'Эквадор'),
    ('EG', 'Egypt', 'Египет'),
    ('SV', 'El Salvador', 'Сальвадор'),
    ('GQ', 'Equatorial Guinea', 'Экваториальная Гвинея'),
    ('ER', 'Eritrea', 'Эритрея'),
    ('EE', 'Estonia', 'Эстония'),
    ('ET', 'Ethiopia', 'Эфиопия'),

    ('FI', 'Finland', 'Финляндия'),
    ('FR', 'France', 'Франция'),

    ('GA', 'Gabon', 'Габон'),
    ('GM', 'Gambia', 'Гамбия'),
    ('GE', 'Georgia', 'Грузия'),
    ('DE', 'Germany', 'Германия'),
    ('GH', 'Ghana', 'Гана'),
    ('GI', 'Gibraltar', 'Гибралтар'),
    ('GR', 'Greece', 'Греция'),
    ('GL', 'Greenland', 'Гренландия'),
    ('GD', 'Grenada', 'Гренада'),
    ('GU', 'Guam', 'Гуам'),
    ('GT', 'Guatemala', 'Гватемала'),
    ('GN', 'Guinea', 'Гвинея'),
    ('GW', 'Guinea-Bissau', 'Гвинея-Бисау'),
    ('GY', 'Guyana', 'Гайана'),

    ('HT', 'Haiti', 'Гаити'),
    ('HN', 'Honduras', 'Гондурас'),
    ('HK', 'Hong Kong', 'Гонконг'),
    ('HU', 'Hungary', 'Венгрия'),

    ('IS', 'Iceland', 'Исландия'),
    ('IN', 'India', 'Индия'),
    ('ID', 'Indonesia', 'Индонезия'),
    ('IR', 'Iran', 'Иран'),
    ('IQ', 'Iraq', 'Ирак'),
    ('IE', 'Ireland', 'Ирландия'),
    ('IL', 'Israel', 'Израиль'),
    ('IT', 'Italy', 'Италия'),

    ('JM', 'Jamaica', 'Ямайка'),
    ('JP', 'Japan', 'Япония'),
    ('JO', 'Jordan', 'Иордания'),

    ('KZ', 'Kazakhstan', 'Казахстан'),
    ('KE', 'Kenya', 'Кения'),
    ('KI', 'Kiribati', 'Кирибати'),
    ('KW', 'Kuwait', 'Кувейт'),
    ('KG', 'Kyrgyzstan', 'Киргизия'),

    ('LA', 'Laos', 'Лаос'),
    ('LV', 'Latvia', 'Латвия'),
    ('LB', 'Lebanon', 'Ливан'),
    ('LS', 'Lesotho', 'Лесото'),
    ('LR', 'Liberia', 'Либерия'),
    ('LY', 'Libya', 'Ливия'),
    ('LI', 'Liechtenstein', 'Лихтенштейн'),
    ('LT', 'Lithuania', 'Литва'),
    ('LU', 'Luxembourg', 'Люксембург'),

    ('MO', 'Macao', 'Макао'),
    ('MG', 'Madagascar', 'Мадагаскар'),
    ('MW', 'Malawi', 'Малави'),
    ('MY', 'Malaysia', 'Малайзия'),
    ('MV', 'Maldives', 'Мальдивы'),
    ('ML', 'Mali', 'Мали'),
    ('MT', 'Malta', 'Мальта'),
    ('MH', 'Marshall Islands', 'Маршалловы острова'),
    ('MQ', 'Martinique', 'Мартиника'),
    ('MR', 'Mauritania', 'Мавритания'),
    ('MU', 'Mauritius', 'Маврикий'),
    ('MX', 'Mexico', 'Мексика'),
    ('FM', 'Micronesia', 'Микронезия'),
    ('MD', 'Moldova', 'Молдова'),
    ('MC', 'Monaco', 'Монако'),
    ('MN', 'Mongolia', 'Монголия'),
    ('ME', 'Montenegro', 'Черногория'),
    ('MA', 'Morocco', 'Марокко'),
    ('MZ', 'Mozambique', 'Мозамбик'),

    ('NA', 'Namibia', 'Намибия'),
    ('NP', 'Nepal', 'Непал'),
    ('NL', 'Netherlands', 'Нидерланды'),
    ('NZ', 'New Zealand', 'Новая Зеландия'),
    ('NI', 'Nicaragua', 'Никарагуа'),
    ('NE', 'Niger', 'Нигер'),
    ('NG', 'Nigeria', 'Нигерия'),
    ('NO', 'Norway', 'Норвегия'),

    ('OM', 'Oman', 'Оман'),

    ('PK', 'Pakistan', 'Пакистан'),
    ('PA', 'Panama', 'Панама'),
    ('PY', 'Paraguay', 'Парагвай'),
    ('PE', 'Peru', 'Перу'),
    ('PH', 'Philippines', 'Филиппины'),
    ('PL', 'Poland', 'Польша'),
    ('PT', 'Portugal', 'Португалия'),
    ('PR', 'Puerto Rico', 'Пуэрто-Рико'),

    ('QA', 'Qatar', 'Катар'),

    ('RO', 'Romania', 'Румыния'),
    ('RU', 'Russia', 'Россия'),
    ('RW', 'Rwanda', 'Руанда'),

    ('SA', 'Saudi Arabia', 'Саудовская Аравия'),
    ('SN', 'Senegal', 'Сенегал'),
    ('RS', 'Serbia', 'Сербия'),
    ('SG', 'Singapore', 'Сингапур'),
    ('SK', 'Slovakia', 'Словакия'),
    ('SI', 'Slovenia', 'Словения'),
    ('ZA', 'South Africa', 'ЮАР'),
    ('KR', 'South Korea', 'Южная Корея'),
    ('ES', 'Spain', 'Испания'),
    ('LK', 'Sri Lanka', 'Шри-Ланка'),
    ('SE', 'Sweden', 'Швеция'),
    ('CH', 'Switzerland', 'Швейцария'),

    ('TH', 'Thailand', 'Таиланд'),
    ('TR', 'Turkey', 'Турция'),
    ('TW', 'Taiwan', 'Тайвань'),

    ('UA', 'Ukraine', 'Украина'),
    ('AE', 'United Arab Emirates', 'ОАЭ'),
    ('GB', 'United Kingdom', 'Великобритания'),
    ('US', 'United States of America', 'США'),
    ('UY', 'Uruguay', 'Уругвай'),

    ('VE', 'Venezuela', 'Венесуэла'),
    ('VN', 'Vietnam', 'Вьетнам'),

    ('ZM', 'Zambia', 'Замбия'),
    ('ZW', 'Zimbabwe', 'Зимбабве');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
