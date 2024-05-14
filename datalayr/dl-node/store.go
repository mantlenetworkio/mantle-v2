package dln

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Store is a database to store frames and a logger to get domain level logs
type Store struct {
	db     *LevelDBStore
	logger *logging.Logger
}

// KeyPair is a
type KeyPair struct {
	Id     uint64
	Commit []byte
}

// NewStore creates a new Store object with a db at the provided path and the given logger
func NewStore(path string, logger *logging.Logger) (*Store, error) {
	logger.Trace().Msg("Entering NewStore function...")
	defer logger.Trace().Msg("Exiting NewStore function...")

	// create the db at the path
	db, err := NewLevelDBStore(path)
	if err != nil {
		logger.Error().Err(err).Msg("Could not create leveldb database")
		return nil, err
	}

	return &Store{
		db:     db,
		logger: logger,
	}, nil
}

// UpdateLatestBlockTime returns the latest expiry time among all the currently stored datatstores
func (s *Store) GetLatestBlockTime() (uint64, bool) {
	key := []byte("LatestBlockTime")
	data, err := s.db.Get(key)
	if err != nil {
		return 0, false
	}
	bn := toUint64(data)
	return bn, true
}

// UpdateLatestBlockTime updates the latest expiry time among all the currently stored datatstores
func (s *Store) UpdateLatestBlockTime(bn uint64) bool {
	key := []byte("LatestBlockTime")
	data := toByteArray(bn)
	err := s.db.Put(key, data)
	return err == nil
}

// InsertCommit inserts a serialized list of byte arrays into the db with the key being a sepcified commit
//
// returns an error if serializing the list or insertion into the db fails
func (s *Store) InsertCommit(ctx context.Context, commit []byte, frames [][]byte) error {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering StoreFrames function...")
	defer log.Trace().Msg("Exiting StoreFrames function...")

	_, err := s.db.Get(commit)
	if err == nil {
		log.Error().Err(ErrKeyAlreadyExists)
		return ErrKeyAlreadyExists
	}

	data, err := encodeFrames(frames)
	if err != nil {
		log.Error().Err(err).Msg("Could not encode frames")
		return err
	}

	log.Debug().Msgf("data has length %v", len(data))

	err = s.db.Put(commit, data)
	if err != nil {
		log.Error().Err(err).Msg("Could not store frame in db")
		return err
	}

	return nil
}

// HashCommit returns if a given commit has been stored, true with existing commits
func (s *Store) HasCommit(ctx context.Context, commit []byte) bool {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering StoreFrames function...")
	defer log.Trace().Msg("Exiting StoreFrames function...")

	_, err := s.db.Get(commit)
	return err == nil
}

// GetFrames returns the list of byte arrays stored for given commit along with a boolean if
// the read was unsuccessful or the frames were not serialized correctly
func (s *Store) GetFrames(ctx context.Context, commit []byte) ([][]byte, bool) {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering GetFrames function...")
	defer log.Trace().Msg("Exiting GetFrames function...")

	data, err := s.db.Get(commit)
	if err != nil {
		return nil, false
	}
	log.Info().Msgf("Get frame %v has length %v", hexutil.Encode(commit), len(data))

	frames, err := decodeFrames(data)
	if err != nil {
		return nil, false
	}
	return frames, true

}

// Expire expires a commit by removing the frame from the db
func (s *Store) Expire(commit []byte) error {
	return s.db.Delete(commit)
}

// encodeFrames flattens an array of byte arrays (frames) into a single byte array
//
// encodeFrames(frames) = (len(frames[0]), frames[0], len(frames[1]), frames[1], ...)
func encodeFrames(frames [][]byte) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	for _, frame := range frames {
		if err := binary.Write(buf, binary.LittleEndian, uint64(len(frame))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(frame); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// DecodeFrames turns a flattened array of frames into an array of its constituent frames
// throwing an error in case the frames were not serialized correctly
//
// DecodeFrames((len(frames[0]), frames[0], len(frames[1]), frames[1], ...)) = frames
func decodeFrames(data []byte) ([][]byte, error) {
	buf := bytes.NewReader(data)
	frames := make([][]byte, 0)

	for {
		var length uint64
		if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
			break
		}
		frame := make([]byte, length)
		buf.Read(frame)
		frames = append(frames, frame)

		if buf.Len() < 8 {
			if buf.Len() != 0 {
			}
			break
		}
	}

	return frames, nil
}
