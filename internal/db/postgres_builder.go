package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"messaging-server/internal/logger"
	"messaging-server/internal/model"
	"os"
)

type postgresTemplate struct {
	connStr string
	DB      *sql.DB
}

var PostgresConnection *postgresTemplate

// InitPostgres sets up Postgres struct and opens the initial connection
func InitPostgres() error {
	PostgresConnection = &postgresTemplate{
		connStr: os.Getenv("POSTGRES_URI"),
	}
	return PostgresConnection.connect()
}

// connect (re)opens p.DB when connection lost
func (p *postgresTemplate) connect() error {
	// control if db conn already exist and alive
	// TODO make function
	if p.DB != nil {
		if err := p.DB.Ping(); err == nil {
			return nil
		} else {
			_ = p.DB.Close()
		}
	}

	// create connection if not exist
	db, err := sql.Open("postgres", p.connStr)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping failed: %w", err)
	}
	p.DB = db

	logger.Sugar.Info("postgres connection established successfully")
	return nil
}

// FetchUnsentMessages get records from Postgres unsent in a list
func (p *postgresTemplate) FetchUnsentMessages(limit int) ([]model.Message, error) {
	if err := p.connect(); err != nil {
		return nil, err
	}

	rows, err := p.DB.Query(`
		SELECT 
		       id,
		       content,
		       phone,
		       isSent
		FROM messages
		WHERE isSent is FALSE
		ORDER BY id ASC
		LIMIT $1`, limit,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var msgs []model.Message

	// loop over the messages
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.Id, &m.Content, &m.Phone, &m.Status); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	// check iteration error
	if err := rows.Err(); err != nil {
		return nil, err
	}

	logger.Sugar.Infof("%d messages fetched", len(msgs))
	return msgs, nil
}

// UpdateSentMessage update the status of sent messages in Postgres
func (p *postgresTemplate) UpdateSentMessage(msg model.Message) error {
	// when try to reconnect
	if err := p.connect(); err != nil {
		return err
	}

	// set true if successfully sent
	if _, err := p.DB.Exec(`
				UPDATE messages
                SET isSent = TRUE
                WHERE id = $1`,
		msg.Id,
	); err != nil {
		return err
	}

	logger.Sugar.Infof("message %d updated", msg.Id)

	return nil
}

// GetSentMessages list messages sent
func (p *postgresTemplate) GetSentMessages() ([]model.Message, error) {
	// when try to reconnect
	if err := p.connect(); err != nil {
		return nil, err
	}

	// get all info from postgres that records sent
	rows, err := p.DB.Query(
		`SELECT 
    				id,
    				content,
    				phone,
    				isSent
    			FROM messages
             	WHERE isSent is TRUE
             	LIMIT 10`,
	)

	if err != nil {
		return nil, err
	}

	var msgs []model.Message
	// iterate and scan each row into message struct
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.Id, &m.Content, &m.Phone, &m.Status); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	// check iteration error
	if err := rows.Err(); err != nil {
		return nil, err
	}

	logger.Sugar.Infof("retrieved %d sent messages", len(msgs))
	return msgs, nil
}
