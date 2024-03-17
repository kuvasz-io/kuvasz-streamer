package main

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
)

type Worker struct {
	log         *slog.Logger
	workChannel chan operation
	jobsCounter prometheus.Counter
	tx          pgx.Tx
}

var Workers []Worker

func (w Worker) work() {
	var op operation
	var err error
	log := w.log
	timer := time.NewTimer(time.Duration(config.App.CommitDelay) * time.Second)
	for {
		select {
		case <-timer.C:
			log.Debug("timer expired, committing transaction")
			if w.tx == nil {
				log.Debug("no transaction to commit, restarting timer")
				timer.Reset(time.Duration(config.App.CommitDelay) * time.Second)
				continue
			}
			if err = w.tx.Commit(context.Background()); err != nil {
				log.Error("failed to commit transaction", "error", err)
			}
			log.Debug("committed transaction")
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
		go Workers[i].work()
	}
}
