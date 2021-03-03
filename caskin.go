package caskin

type Caskin struct {
	options *Options
}

func (c *Caskin) GetExecutor(provider CurrentProvider) *executor {
	e := NewEnforcer(c.options.Enforcer, c.options.EntryFactory)
	return &executor{
		e:        e,
		factory:  c.options.EntryFactory,
		mdb:      c.options.MetaDB,
		provider: provider,
		options:  c.options,
	}
}

func New(options *Options, opts ...Option) (*Caskin, error) {
	options = options.newOptions(opts...)
	if options.DomainCreator == nil {
		return nil, ErrInitializationNilDomainCreator
	}
	if options.Enforcer == nil {
		return nil, ErrInitializationNilEnforcer
	}
	if options.EntryFactory == nil {
		return nil, ErrInitializationNilEntryFactory
	}
	if options.MetaDB == nil {
		return nil, ErrInitializationNilMetaDB
	}
	return &Caskin{options: options}, nil
}
