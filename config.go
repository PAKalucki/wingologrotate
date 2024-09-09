package main

type LogEntry struct {
	Path      string     `yaml:"path"`
	Type      string     `yaml:"type"`
	Schedule  string     `yaml:"schedule,omitempty"`
	Size      string     `yaml:"size,omitempty"`
	MaxKeep   int        `yaml:"max_keep,omitempty"`
	Condition *Condition `yaml:"condition,omitempty"`
}

type Condition struct {
	Age string `yaml:"age,omitempty"`
}

type Config struct {
	Logs     []LogEntry `yaml:"logs"`
	Schedule string     `yaml:"schedule"`
}
