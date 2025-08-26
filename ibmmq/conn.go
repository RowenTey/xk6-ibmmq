package ibmmq

import "github.com/ibm-messaging/mq-golang-jms20/jms20subset"

func (p *Ibmmq) commit(ctx jms20subset.JMSContext) jms20subset.JMSException {
	return ctx.Commit()
}

func (p *Ibmmq) close(ctx jms20subset.JMSContext) {
	ctx.Close()
}
