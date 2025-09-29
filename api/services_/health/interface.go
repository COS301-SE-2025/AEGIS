package health


type Checker interface {
	CheckPostgres() ComponentStatus
	CheckMongo() ComponentStatus
	CheckIPFS() ComponentStatus
	CheckDisk() ComponentStatus
	CheckMemory() ComponentStatus
}
