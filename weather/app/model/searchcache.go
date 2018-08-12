package model

import (
	"fmt"
	"time"

	"../shared/database"
)

// *****************************************************************************
// search_cache
// *****************************************************************************

// SearchCache table contains the information for each user
type SearchCache struct {
	//ObjectID  bson.ObjectId `bson:"_id"`
	//ID        uint32        `db:"id" bson:"id,omitempty"` // Don't use Id, use UserID() instead for consistency with MongoDB
	Zipcode   string    `db:"zipcode" bson:"zipcode"`
	City      string    `db:"city" bson:"city"`
	State     string    `db:"state" bson:"state"`
	TimesUsed uint32    `db:"times_used" bson:"times_used"`
	CurrTemp  string    `db:"currtemp" bson:"currtemp"`
	HighTemp  string    `db:"hightemp" bson:"hightemp"`
	LowTemp   string    `db:"lowtemp" bson:"lowtemp"`
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`
	Deleted   uint8     `db:"deleted" bson:"deleted"`
}

// ZipcodeID returns the zipcode
func (u *SearchCache) ZipcodeID() string {
	r := ""

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		r = fmt.Sprintf("%v", u.Zipcode)
		/*
			case database.TypeMongoDB:
				r = u.ObjectID.Hex()
			case database.TypeBolt:
				r = u.ObjectID.Hex()
		*/
	}

	return r
}

// SearchCacheCreate creates user
func SearchCacheCreate(_sc SearchCache) error {
	var err error

	now := time.Now()

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO search_cache (zipcode, times_used, currtemp, hightemp, lowtemp, updated_at) VALUES (?,?,?,?,?,?)", _sc.Zipcode, 1, _sc.CurrTemp, _sc.HighTemp, _sc.LowTemp, now)
		/*
			case database.TypeMongoDB:
				if database.CheckConnection() {
					session := database.Mongo.Copy()
					defer session.Close()
					c := session.DB(database.ReadConfig().MongoDB.Database).C("search_cache")

					user := &User{
						ObjectID:  bson.NewObjectId(),
						FirstName: firstName,
						LastName:  lastName,
						Email:     email,
						Password:  password,
						StatusID:  1,
						CreatedAt: now,
						UpdatedAt: now,
						Deleted:   0,
					}
					err = c.Insert(user)
				} else {
					err = ErrUnavailable
				}
			case database.TypeBolt:
				user := &User{
					ObjectID:  bson.NewObjectId(),
					FirstName: firstName,
					LastName:  lastName,
					Email:     email,
					Password:  password,
					StatusID:  1,
					CreatedAt: now,
					UpdatedAt: now,
					Deleted:   0,
				}

				err = database.Update("user", user.Email, &user)
		*/
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// SearchCacheByZipcode gets note by Zipcode
func SearchCacheByZipcode(_zipcode string) (SearchCache, error) {
	var err error

	result := SearchCache{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT zipcode, times_used, currtemp, hightemp, lowtemp, updated_at, deleted FROM search_cache WHERE zipcode = ? LIMIT 1", _zipcode)
		/*
			case database.TypeMongoDB:
				if database.CheckConnection() {
					// Create a copy of mongo
					session := database.Mongo.Copy()
					defer session.Close()
					c := session.DB(database.ReadConfig().MongoDB.Database).C("note")

					// Validate the object id
					if bson.IsObjectIdHex(noteID) {
						err = c.FindId(bson.ObjectIdHex(noteID)).One(&result)
						if result.UserID != bson.ObjectIdHex(userID) {
							result = Note{}
							err = ErrUnauthorized
						}
					} else {
						err = ErrNoResult
					}
				} else {
					err = ErrUnavailable
				}
			case database.TypeBolt:
				err = database.View("note", userID+noteID, &result)
				if err != nil {
					err = ErrNoResult
				}
				if result.UserID != bson.ObjectIdHex(userID) {
					result = Note{}
					err = ErrUnauthorized
				}
		*/
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// SearchCacheByAll gets all search_cache
func SearchCacheByAll() ([]SearchCache, error) {
	var err error

	var result []SearchCache

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result, "SELECT zipcode, times_used, currtemp, hightemp, lowtemp, updated_at, deleted FROM search_cache WHERE deleted = 0")
		/*
			case database.TypeMongoDB:
				if database.CheckConnection() {
					// Create a copy of mongo
					session := database.Mongo.Copy()
					defer session.Close()
					c := session.DB(database.ReadConfig().MongoDB.Database).C("note")

					// Validate the object id
					if bson.IsObjectIdHex(userID) {
						err = c.Find(bson.M{"user_id": bson.ObjectIdHex(userID)}).All(&result)
					} else {
						err = ErrNoResult
					}
				} else {
					err = ErrUnavailable
				}
			case database.TypeBolt:
				// View retrieves a record set in Bolt
				err = database.BoltDB.View(func(tx *bolt.Tx) error {
					// Get the bucket
					b := tx.Bucket([]byte("note"))
					if b == nil {
						return bolt.ErrBucketNotFound
					}

					// Get the iterator
					c := b.Cursor()

					prefix := []byte(userID)
					for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
						var single Note

						// Decode the record
						err := json.Unmarshal(v, &single)
						if err != nil {
							log.Println(err)
							continue
						}

						result = append(result, single)
					}

					return nil
				})
		*/
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// SearchCacheUpdate updates a search_cache
func SearchCacheUpdate(_zipcode string, _currtemp string, _hightemp string, _lowtemp string, _deleted uint16) error {
	var err error

	now := time.Now()

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE search_cache SET times_used = times_used+1, currtemp=?, hightemp=?, lowtemp=?, updated_at=?, deleted=? WHERE zipcode = ? LIMIT 1", _currtemp, _hightemp, _lowtemp, now, _deleted, _zipcode)
		/*
			case database.TypeMongoDB:
				if database.CheckConnection() {
					// Create a copy of mongo
					session := database.Mongo.Copy()
					defer session.Close()
					c := session.DB(database.ReadConfig().MongoDB.Database).C("note")
					var note Note
					note, err = NoteByID(userID, noteID)
					if err == nil {
						// Confirm the owner is attempting to modify the note
						if note.UserID.Hex() == userID {
							note.UpdatedAt = now
							note.Content = content
							err = c.UpdateId(bson.ObjectIdHex(noteID), &note)
						} else {
							err = ErrUnauthorized
						}
					}
				} else {
					err = ErrUnavailable
				}
			case database.TypeBolt:
				var note Note
				note, err = NoteByID(userID, noteID)
				if err == nil {
					// Confirm the owner is attempting to modify the note
					if note.UserID.Hex() == userID {
						note.UpdatedAt = now
						note.Content = content
						err = database.Update("note", userID+note.ObjectID.Hex(), &note)
					} else {
						err = ErrUnauthorized
					}
				}
		*/
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// SearchCacheDelete deletes a search_cache
func SearchCacheDelete(_zipcode string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM search_cache WHERE zipcode = ?", _zipcode)
		/*
			case database.TypeMongoDB:
				if database.CheckConnection() {
					// Create a copy of mongo
					session := database.Mongo.Copy()
					defer session.Close()
					c := session.DB(database.ReadConfig().MongoDB.Database).C("note")

					var note Note
					note, err = NoteByID(userID, noteID)
					if err == nil {
						// Confirm the owner is attempting to modify the note
						if note.UserID.Hex() == userID {
							err = c.RemoveId(bson.ObjectIdHex(noteID))
						} else {
							err = ErrUnauthorized
						}
					}
				} else {
					err = ErrUnavailable
				}
			case database.TypeBolt:
				var note Note
				note, err = NoteByID(userID, noteID)
				if err == nil {
					// Confirm the owner is attempting to modify the note
					if note.UserID.Hex() == userID {
						err = database.Delete("note", userID+note.ObjectID.Hex())
					} else {
						err = ErrUnauthorized
					}
				}
		*/
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
