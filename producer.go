package kafka

type Producer struct {
	client            *Client
	topic             string
	partitioner       PartitionChooser
	responseCondition int16
	responseTimeout   int32
}

func NewProducer(client *Client, topic string, partitioner PartitionChooser, responseCondition int16, responseTimeout int32) *Producer {
	return &Producer{client, topic, partitioner, responseCondition, responseTimeout}
}

func NewSimpleProducer(client *Client, topic string) *Producer {
	return NewProducer(client, topic, RandomPartitioner{}, WAIT_FOR_LOCAL, 0)
}

func (p *Producer) SendMessage(key, value encoder) error {
	partitions, err := p.client.brokers.partitionsForTopic(p.topic)
	if err != nil {
		return err
	}

	var partitioner PartitionChooser
	if key == nil {
		partitioner = RandomPartitioner{}
	} else {
		partitioner = p.partitioner
	}
	partition := partitioner.ChoosePartition(key, partitions)

	msg, err := newMessage(key, value)
	if err != nil {
		return err
	}

	request := newSingletonProduceRequest(p.topic, partition, newSingletonMessageSet(msg))
	request.requiredAcks = p.responseCondition
	request.timeout = p.responseTimeout

	_, err = p.client.brokers.sendToPartition(p.topic, partition, request, &produceResponse{})

	return err
}

type encodableString string

func (s encodableString) encode(pe packetEncoder) {
	pe.putRaw([]byte(s))
}

func (p *Producer) SendSimpleMessage(in string) error {
	return p.SendMessage(nil, encodableString(in))
}
