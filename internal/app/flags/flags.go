package flags

import "flag"

const NotAvailable string = "N/A"

type Flags struct {
	C   string
	A   string
	B   string
	F   string
	D   string
	S   string
	Crt string
	Key string
}

func NewFlags() *Flags {
	flags := &Flags{}

	flag.StringVar(&flags.C, "c", "", "Файл конфигурации")
	flag.StringVar(&flags.A, "a", NotAvailable, "Адрес запуска HTTP-сервера")
	flag.StringVar(&flags.B, "b", NotAvailable, "Базовый адрес результирующего сокращённого URL")
	flag.StringVar(&flags.F, "f", NotAvailable, "Путь до файла с сокращёнными URL")
	flag.StringVar(&flags.D, "d", NotAvailable, "Строка с адресом подключения к БД")
	flag.StringVar(&flags.S, "s", NotAvailable, "Использовать HTTPS")
	flag.StringVar(&flags.Crt, "crt", NotAvailable, "Файл с сертификатом")
	flag.StringVar(&flags.Key, "key", NotAvailable, "Файл с приватным ключом")
	flag.Parse()

	return flags
}
