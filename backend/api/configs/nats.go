package configs

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
)

type NATSQueue struct {
	Conn           *nats.Conn
	JetStream      nats.JetStreamContext
	Stream         string
	Subject        string
	Consumer       string
	FetchBatchSize int
	FetchMaxWait   time.Duration
	MaxAckPending  int
}

func ConnectNATS(env *Env) (*NATSQueue, error) {
	nc, err := nats.Connect(env.NATSURL)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}

	if err := ensureStream(js, env.NATSStream, env.NATSSubject); err != nil {
		nc.Close()
		return nil, err
	}

	if err := ensureConsumer(js, env.NATSStream, env.NATSConsumer, env.NATSSubject, env.NATSMaxAckPending); err != nil {
		nc.Close()
		return nil, err
	}

	return &NATSQueue{
		Conn:           nc,
		JetStream:      js,
		Stream:         env.NATSStream,
		Subject:        env.NATSSubject,
		Consumer:       env.NATSConsumer,
		FetchBatchSize: env.NATSFetchBatchSize,
		FetchMaxWait:   time.Duration(env.NATSFetchMaxWaitMS) * time.Millisecond,
		MaxAckPending:  env.NATSMaxAckPending,
	}, nil
}

func (q *NATSQueue) Close() {
	if q == nil || q.Conn == nil {
		return
	}
	q.Conn.Drain()
	q.Conn.Close()
}

func (q *NATSQueue) Publish(ctx context.Context, payload []byte) error {
	if q == nil || q.JetStream == nil {
		return errors.New("nats queue is not configured")
	}
	_, err := q.JetStream.Publish(q.Subject, payload, nats.Context(ctx))
	return err
}

// Broadcast sends a lightweight, fire-and-forget message via NATS Core (for Live Tail)
func (q *NATSQueue) Broadcast(subject string, payload []byte) error {
	if q == nil || q.Conn == nil {
		return errors.New("nats core connection is not configured")
	}
	return q.Conn.Publish(subject, payload)
}

func (q *NATSQueue) PullSubscribe() (*nats.Subscription, error) {
	if q == nil || q.JetStream == nil {
		return nil, errors.New("nats queue is not configured")
	}
	return q.JetStream.PullSubscribe(q.Subject, q.Consumer, nats.BindStream(q.Stream))
}

func ensureStream(js nats.JetStreamContext, streamName string, subject string) error {
	if _, err := js.StreamInfo(streamName); err == nil {
		return nil
	}

	_, err := js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{subject},
		Retention: nats.LimitsPolicy,
		Storage:   nats.FileStorage,
		Discard:   nats.DiscardOld,
		MaxAge:    7 * 24 * time.Hour,
	})
	return err
}

func ensureConsumer(js nats.JetStreamContext, streamName string, consumerName string, subject string, maxAckPending int) error {
	config := &nats.ConsumerConfig{
		Durable:       consumerName,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       5 * time.Minute,
		MaxAckPending: maxAckPending,
		FilterSubject: subject,
		ReplayPolicy:  nats.ReplayInstantPolicy,
	}

	// Update the mutable flow-control setting as well as creating a new
	// consumer. Without this, changing an environment variable would have no
	// effect after the durable consumer had been created once.
	if existing, err := js.ConsumerInfo(streamName, consumerName); err == nil {
		existing.Config.MaxAckPending = maxAckPending
		_, err = js.UpdateConsumer(streamName, &existing.Config)
		return err
	}

	_, err := js.AddConsumer(streamName, config)
	return err
}
