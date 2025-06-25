package case_creation

type CaseRepository interface {
	CreateCase(c *Case) error
}
