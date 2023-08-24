package illacapacity

const (
	CAPACITY_INSTANCE_TYPE_TEAM_LICENSE                    = 1
	CAPACITY_INSTANCE_TYPE_DRIVE_VOLUME                    = 2
	CAPACITY_INSTANCE_TYPE_DRIVE_TRAFFIC                   = 3
	CAPACITY_INSTANCE_TYPE_POSTGRES_DATABASE_RECORD_VOLUME = 4
	CAPACITY_INSTANCE_TYPE_TEAM_LICENSE_APPSUMO            = 5
	CAPACITY_INSTANCE_TYPE_COLLA                           = 6
)

// virtual instance default id
const (
	TEAM_LICENSE_DEFAULT_INSTANCE_ID = -1 // because team license is a virtual instance (have no database storage handle it)
	COLLA_DEFAULT_INSTANCE_ID        = -2 // because colla is a virtual instance (have no database storage handle it)
)
