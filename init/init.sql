
GRANT ALL PRIVILEGES ON DATABASE mydb TO myuser;

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY ,
    content VARCHAR(200) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    isSent  BOOLEAN NOT NULL DEFAULT FALSE
);

-- dummy
INSERT INTO messages (content, phone) VALUES
('Hello, this is the 1st message', '+15551230001'),
('Second message here',           '+15551230002'),
('Third message content',         '+15551230003'),
('Fourth one coming through',     '+15551230004'),
('Fifth and final message',       '+15551230005'),
('Sixth message arriving now',     '+15551230006'),
('Seventh message inbound',         '+15551230007'),
('Eighth message rolling out',      '+15551230008'),
('Ninth message just sent',         '+15551230009'),
('Tenth message check in',          '+15551230010'),
('Eleventh message live',           '+15551230011'),
('Twelfth message update',          '+15551230012'),
('Thirteenth message delivered',    '+15551230013'),
('Fourteenth message broadcast',    '+15551230014'),
('Fifteenth and final message here','+15551230015');
