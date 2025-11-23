CREATE TABLE consumer_otps (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    otp VARCHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    used BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_consumer_otps_email ON consumer_otps(email);
CREATE INDEX idx_consumer_otps_email_otp ON consumer_otps(email, otp);

