package main

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	lsnStatus struct {
		ConfirmedLSN pglogrepl.LSN
		WrittenLSN   pglogrepl.LSN
		CommittedLSN pglogrepl.LSN
	}
	sourceStatus struct {
		sync.Mutex
		m map[string]lsnStatus
	}

	Worker struct {
		log         *slog.Logger
		workChannel chan operation
		jobsCounter prometheus.Counter
		tx          pgx.Tx
		s           *sourceStatus
	}
)

var Workers []Worker

func (s *sourceStatus) Write(dbsid string, lsn pglogrepl.LSN) {
	s.Lock()
	defer s.Unlock()
	if s.m == nil {
		s.m = make(map[string]lsnStatus)
	}
	status := s.m[dbsid]
	status.WrittenLSN = lsn
	s.m[dbsid] = status
}

func (s *sourceStatus) Commit() {
	s.Lock()
	defer s.Unlock()
	if s.m == nil {
		log.Error("commit with no write")
		return
	}
	for k, v := range s.m {
		v.CommittedLSN = v.WrittenLSN
		s.m[k] = v
	}
}

func (w Worker) work() {
	var op operation
	var err error
	log := w.log
	timer := time.NewTimer(time.Duration(config.App.CommitDelay) * time.Second)
	for {
		select {
		case <-timer.C:
			if w.tx == nil {
				timer.Reset(time.Duration(config.App.CommitDelay) * time.Second)
				continue
			}
			if err = w.tx.Commit(context.Background()); err != nil {
				log.Error("failed to commit transaction", "error", err)
			}
			w.s.Commit()
			log.Debug("committed transaction", "lsn", w.s)
			w.tx = nil
			timer.Reset(time.Duration(config.App.CommitDelay) * time.Second)
		case op = <-w.workChannel:
			w.jobsCounter.Inc()
			if w.tx == nil {
				w.tx, err = DestConnectionPool.Begin(context.Background())
				if err != nil {
					log.Error("failed to begin transaction", "error", err)
					continue
				}
			}
			log.Debug("received operation", "op", op)
			switch op.opCode {
			case "ic":
				_ = op.insertClone(w.tx)
			case "uc":
				_ = op.updateClone(w.tx)
			case "dc":
				_ = op.deleteClone(w.tx)
			default:
				log.Error("unhandled opcode", "op", op.opCode)
			}
			w.s.Write(op.database+"-"+op.sid, op.lsn)
			log.Debug("Performed operation", "op", op, "lsn", w.s)
		}
	}
}

func SendWork(op operation) {
	Workers[op.id%int64(len(Workers))].workChannel <- op
}

func StartWorkers(numWorkers int) {
	Workers = make([]Worker, numWorkers)
	for i := 0; i < numWorkers; i++ {
		Workers[i].workChannel = make(chan operation)
		Workers[i].jobsCounter = jobsTotal.WithLabelValues(strconv.Itoa(i))
		Workers[i].log = log.With("worker", i)
		Workers[i].s = &sourceStatus{m: make(map[string]lsnStatus)}
		Workers[i].tx = nil
		go Workers[i].work()
	}
}

func SetCommittedLSN(database, sid string, lsn pglogrepl.LSN) {
	dbsid := database + "-" + sid

	for i := range Workers {
		status := Workers[i].s.m[dbsid]
		status.CommittedLSN = lsn
		Workers[i].s.m[dbsid] = status
	}
}

func GetCommittedLSN(database, sid string, sourceCommittedLSN pglogrepl.LSN) pglogrepl.LSN {
	dbsid := database + "-" + sid
	// log := log.With("db-sid", dbsid)
	lowestDirtyLSN := pglogrepl.LSN(0)
	highestCommittedLSN := pglogrepl.LSN(0)

	// step 1 find lowest dirty LSN
	for i := range Workers {
		if status, ok := Workers[i].s.m[dbsid]; ok {
			if status.WrittenLSN > status.CommittedLSN { // worker has written requests but not committed
				if status.WrittenLSN < lowestDirtyLSN || lowestDirtyLSN == 0 {
					lowestDirtyLSN = status.WrittenLSN
				}
			}
		}
	}
	// log.Debug("Found Lowest dirty LSN", "lowestDirtyLSN", lowestDirtyLSN)
	// step 2 find highest committed transaction in destination already committed in the source
	for i := range Workers {
		// log.Debug("Worker info", "i", i, "m", Workers[i].s.m[dbsid])
		if status, ok := Workers[i].s.m[dbsid]; ok {
			if (status.CommittedLSN < lowestDirtyLSN || lowestDirtyLSN == 0) &&
				status.CommittedLSN <= sourceCommittedLSN &&
				status.CommittedLSN > highestCommittedLSN {
				highestCommittedLSN = status.CommittedLSN
			}
		}
	}
	// log.Debug("Found highest committed LSN", "highestCommittedLSN", highestCommittedLSN, "sourceCommittedLSN", sourceCommittedLSN)
	return highestCommittedLSN
}
