package ignition

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/coreos/go-systemd/unit"
	"github.com/coreos/ignition/config/v2_1/types"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// globalCache keeps the instances of the internal types of ignition generated
// by the different data resources with the goal to be reused by the
// ignition_config data resource. The key of the maps are a hash of the types
// calculated on the type serialized to JSON.
var globalCache = &cache{
	disks:         make(map[string]*types.Disk, 0),
	arrays:        make(map[string]*types.Raid, 0),
	filesystems:   make(map[string]*types.Filesystem, 0),
	files:         make(map[string]*types.File, 0),
	systemdUnits:  make(map[string]*types.Unit, 0),
	networkdUnits: make(map[string]*types.Networkdunit, 0),
	users:         make(map[string]*types.PasswdUser, 0),
	groups:        make(map[string]*types.PasswdGroup, 0),
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"ignition_config":        resourceConfig(),
			"ignition_disk":          resourceDisk(),
			"ignition_raid":          resourceRaid(),
			"ignition_filesystem":    resourceFilesystem(),
			"ignition_file":          resourceFile(),
			"ignition_systemd_unit":  resourceSystemdUnit(),
			"ignition_networkd_unit": resourceNetworkdUnit(),
			"ignition_user":          resourceUser(),
			"ignition_group":         resourceGroup(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"ignition_config": schema.DataSourceResourceShim(
				"ignition_config",
				resourceConfig(),
			),
			"ignition_disk": schema.DataSourceResourceShim(
				"ignition_disk",
				resourceDisk(),
			),
			"ignition_raid": schema.DataSourceResourceShim(
				"ignition_raid",
				resourceRaid(),
			),
			"ignition_filesystem": schema.DataSourceResourceShim(
				"ignition_filesystem",
				resourceFilesystem(),
			),
			"ignition_file": schema.DataSourceResourceShim(
				"ignition_file",
				resourceFile(),
			),
			"ignition_systemd_unit": schema.DataSourceResourceShim(
				"ignition_systemd_unit",
				resourceSystemdUnit(),
			),
			"ignition_networkd_unit": schema.DataSourceResourceShim(
				"ignition_networkd_unit",
				resourceNetworkdUnit(),
			),
			"ignition_user": schema.DataSourceResourceShim(
				"ignition_user",
				resourceUser(),
			),
			"ignition_group": schema.DataSourceResourceShim(
				"ignition_group",
				resourceGroup(),
			),
		},
	}
}

type cache struct {
	disks         map[string]*types.Disk
	arrays        map[string]*types.Raid
	filesystems   map[string]*types.Filesystem
	files         map[string]*types.File
	systemdUnits  map[string]*types.Unit
	networkdUnits map[string]*types.Networkdunit
	users         map[string]*types.PasswdUser
	groups        map[string]*types.PasswdGroup

	sync.Mutex
}

func (c *cache) addDisk(g *types.Disk) string {
	c.Lock()
	defer c.Unlock()

	id := id(g)
	c.disks[id] = g

	return id
}

func (c *cache) addRaid(r *types.Raid) string {
	c.Lock()
	defer c.Unlock()

	id := id(r)
	c.arrays[id] = r

	return id
}

func (c *cache) addFilesystem(f *types.Filesystem) string {
	c.Lock()
	defer c.Unlock()

	id := id(f)
	c.filesystems[id] = f

	return id
}

func (c *cache) addFile(f *types.File) string {
	c.Lock()
	defer c.Unlock()

	id := id(f)
	c.files[id] = f

	return id
}

func (c *cache) addSystemdUnit(u *types.Unit) string {
	c.Lock()
	defer c.Unlock()

	id := id(u)
	c.systemdUnits[id] = u

	return id
}

func (c *cache) addNetworkdUnit(u *types.Networkdunit) string {
	c.Lock()
	defer c.Unlock()

	id := id(u)
	c.networkdUnits[id] = u

	return id
}

func (c *cache) addUser(u *types.PasswdUser) string {
	c.Lock()
	defer c.Unlock()

	id := id(u)
	c.users[id] = u

	return id
}

func (c *cache) addGroup(g *types.PasswdGroup) string {
	c.Lock()
	defer c.Unlock()

	id := id(g)
	c.groups[id] = g

	return id
}

func id(input interface{}) string {
	b, _ := json.Marshal(input)
	return hash(string(b))
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func castSliceInterface(i []interface{}) []string {
	var o []string
	for _, value := range i {
		if value == nil {
			continue
		}

		o = append(o, value.(string))
	}

	return o
}

func getInt(d *schema.ResourceData, key string) *int {
	var i *int
	if value, ok := d.GetOk(key); ok {
		n := value.(int)
		i = &n
	}

	return i
}

var errEmptyUnit = fmt.Errorf("invalid or empty unit content")

func validateUnitContent(content string) error {
	c := bytes.NewBufferString(content)
	unit, err := unit.Deserialize(c)
	if err != nil {
		return fmt.Errorf("invalid unit content: %s", err)
	}

	if len(unit) == 0 {
		return errEmptyUnit
	}

	return nil
}
