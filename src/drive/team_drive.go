package drive

import (
	"strings"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/config"
)

/**
 * team drive folder design
 *
 * - /{bucket-name}
 *   - /team-{team_uid}
 *     - /system  (for team system object storage)
 *       - /icon
 *     - /team (for team drive data)
 */

const TEAM_FOLDER_PREFIX = "team-"
const TEAM_SYSTEM_FOLDER = "/system"
const TEAM_ICON_FOLDER = "/icon"
const TEAM_SPACE_FOLDER = "/team"
const DIGITALOCEAN_REPLACE_TARGET_FOR_TEAM = "sfo3.digitaloceanspaces.com/"

type TeamDrive struct {
	UID              uuid.UUID  `json:"uid"`
	Drive            S3Instance `json:"-"`
	TeamSystemFolder string     `json:"teamsystemfolder"`
	TeamSpaceFolder  string     `json:"teamspacefolder"`
}

func NewTeamDrive(drive *Drive) *TeamDrive {
	return &TeamDrive{
		Drive: drive.TeamDriveS3Instance,
	}
}

func (d *TeamDrive) SetTeam(team *model.Team) {
	d.UID = team.ExportUID()
	d.TeamSystemFolder = TEAM_FOLDER_PREFIX + team.ExportUIDInString() + TEAM_SYSTEM_FOLDER
	d.TeamSpaceFolder = TEAM_FOLDER_PREFIX + team.ExportUIDInString() + TEAM_SPACE_FOLDER
}

func (d *TeamDrive) GetIconUploadPreSignedURL(fileName string) (string, error) {
	path := d.TeamSystemFolder + TEAM_ICON_FOLDER + "/" + fileName
	return d.Drive.GetPreSignedPutURL(path)
}

func (d *TeamDrive) GetAIAgentIconUploadPreSignedURL(fileName string) (string, error) {
	path := d.TeamSystemFolder + TEAM_ICON_FOLDER + "/" + fileName
	return d.Drive.GetPreSignedPutURL(path)
}

func FormatTeamIconURL(url string) string {
	conf := config.GetInstance()
	if conf.DriveType == config.DRIVE_TYPE_DO {
		return strings.Replace(url, DIGITALOCEAN_REPLACE_TARGET_FOR_TEAM, "", -1)
	}
	return url
}
