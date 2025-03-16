package pkg

type Logger interface {
	Log(message string)
}

type ConsoleLogger struct{}

func (cl ConsoleLogger) Log(message string) {
	println(message) // Или используйте более сложные структуры для форматирования
}
