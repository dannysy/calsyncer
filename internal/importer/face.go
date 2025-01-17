package importer

type CalImporter interface {
	Get() (Calendar, error)
}
