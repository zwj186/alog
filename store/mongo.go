package store

import (
	"text/template"

	"github.com/zwj186/alog/log"
	"gopkg.in/mgo.v2"
)

// NewMongoStore 创建基于MongoDB存储的实例
func NewMongoStore(cfg log.MongoConfig) log.LogStore {
	if cfg.URL == "" {
		cfg.URL = log.DefaultMongoURL
	}
	if cfg.DBTmpl == "" {
		cfg.DBTmpl = log.DefaultMongoDBTmpl
	}
	if cfg.CollectionTmpl == "" {
		cfg.CollectionTmpl = log.DefaultMongoCollectionTmpl
	}
	session, err := mgo.Dial(cfg.URL)
	if err != nil {
		panic(err)
	}
	return &MongoStore{
		session:        session,
		isSessionDial:  true,
		cfg:            cfg,
		dbTmpl:         template.Must(template.New("").Parse(cfg.DBTmpl)),
		collectionTmpl: template.Must(template.New("").Parse(cfg.CollectionTmpl)),
	}
}

type MongoStore struct {
	session        *mgo.Session
	isSessionDial  bool
	cfg            log.MongoConfig
	dbTmpl         *template.Template
	collectionTmpl *template.Template
}

func (ms *MongoStore) Store(item *log.LogItem) error {
	if !ms.isSessionDial {
		session, err := mgo.Dial(ms.cfg.URL)
		if err != nil {
			return err
		}
		ms.session = session
		ms.isSessionDial = true
	}
	dbName := log.ParseName(ms.dbTmpl, item)
	collectionName := log.ParseName(ms.collectionTmpl, item)
	err := ms.session.DB(dbName).C(collectionName).Insert(item.ToMap())
	return err
}

func (ms *MongoStore) Close() (err error) {
	ms.session.Close()
	ms.isSessionDial = false
	return err
}
