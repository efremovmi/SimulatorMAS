├── cmd                                
│   └── main.go                  // The entry point to the application
├── config
│   └── config.json              // Application configuration file
├── docker-compose.yml           // Contains RabbitMQ
├── go.mod
├── go.sum
└── internal
    ├── agent
    │   ├── greens
    │   │   └── greens.go        // Contains the implementation of the green agent
    │   └── rodent
    │       └── rodent.go        // Contains an implementation of a rodent agent
    ├── broker
    │   └── broker.go            // Contains the broker implementation
    ├── config_parser
    │   └── config_parser.go     // Config parser
    ├── kernel
    │   ├── kernel.go            // Contains the implementation of the logic of the application core
    │   ├── work_with_greens.go  // Contains the implementation of the logic of the kernel in terms of interaction with the agent greens
    │   └── work_with_rodent.go  // Contains the implementation of the core logic in terms of interaction with the rodent agent
    ├── logger
    │   └── logger.go            // Implements the logging functions of the system
    ├── population
    │   └── population.go        // Wrapper over the creation of agents
    └── types
        ├── igreens.go           // The interface of the agent is green
        ├── ikernel.go           // The interface of the rodent agent
        └── irodent.go           // The interface of the core agent
