package postgres

type GameStorage struct {
	pgConn PgConn
}

func NewGameStorage(pgConn PgConn) (*GameStorage, error) {
	return &GameStorage{
		pgConn: pgConn,
	}, nil
}
