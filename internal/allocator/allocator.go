package allocator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/interpolation"
	"github.com/git-treeline/git-treeline/internal/registry"
)

const maxRedisDbs = 16

type Allocator struct {
	UserConfig    *config.UserConfig
	ProjectConfig *config.ProjectConfig
	Registry      *registry.Registry
}

type Allocation struct {
	Project      string
	Worktree     string
	WorktreeName string
	Port         int
	Ports        []int
	Database     string
	RedisDB      int
	RedisPrefix  string
}

func (a *Allocation) ToRegistryEntry() registry.Allocation {
	entry := registry.Allocation{
		"project":       a.Project,
		"worktree":      a.Worktree,
		"worktree_name": a.WorktreeName,
		"port":          a.Port,
		"ports":         intsToAny(a.Ports),
		"database":      a.Database,
	}

	for i, p := range a.Ports {
		entry[fmt.Sprintf("port_%d", i+1)] = p
	}

	if a.RedisDB > 0 {
		entry["redis_db"] = a.RedisDB
		entry["redis_prefix"] = nil
	} else {
		entry["redis_db"] = nil
		entry["redis_prefix"] = a.RedisPrefix
	}

	return entry
}

func (a *Allocation) ToInterpolationMap() interpolation.Allocation {
	m := interpolation.Allocation{
		"port":          a.Port,
		"ports":         a.Ports,
		"database":      a.Database,
		"worktree_name": a.WorktreeName,
	}
	if a.RedisDB > 0 {
		m["redis_db"] = a.RedisDB
	}
	if a.RedisPrefix != "" {
		m["redis_prefix"] = a.RedisPrefix
	}
	for i, p := range a.Ports {
		m[fmt.Sprintf("port_%d", i+1)] = p
	}
	return m
}

func New(uc *config.UserConfig, pc *config.ProjectConfig, reg *registry.Registry) *Allocator {
	return &Allocator{UserConfig: uc, ProjectConfig: pc, Registry: reg}
}

func (al *Allocator) Allocate(worktreePath, worktreeName string) (*Allocation, error) {
	count := al.ProjectConfig.PortsNeeded()
	if count > al.UserConfig.PortIncrement() {
		return nil, fmt.Errorf("ports_needed (%d) exceeds port.increment (%d); increase port.increment in your config.json to at least %d",
			count, al.UserConfig.PortIncrement(), count)
	}

	ports := al.nextAvailablePorts(count)
	redisDB, redisPrefix := al.allocateRedis(worktreeName)
	database := al.buildDatabaseName(worktreeName)

	return &Allocation{
		Project:      al.ProjectConfig.Project(),
		Worktree:     worktreePath,
		WorktreeName: worktreeName,
		Port:         ports[0],
		Ports:        ports,
		Database:     database,
		RedisDB:      redisDB,
		RedisPrefix:  redisPrefix,
	}, nil
}

func (al *Allocator) BuildRedisURL(alloc *Allocation) string {
	m := alloc.ToInterpolationMap()
	return interpolation.BuildRedisURL(al.UserConfig.RedisURL(), m)
}

func (al *Allocator) nextAvailablePorts(count int) []int {
	usedSet := make(map[int]bool)
	for _, p := range al.Registry.UsedPorts() {
		usedSet[p] = true
	}

	candidate := al.UserConfig.PortBase() + al.UserConfig.PortIncrement()
	for {
		block := make([]int, count)
		conflict := false
		for i := range count {
			port := candidate + i
			block[i] = port
			if usedSet[port] {
				conflict = true
			}
		}
		if !conflict {
			return block
		}
		candidate += al.UserConfig.PortIncrement()
	}
}

func (al *Allocator) allocateRedis(worktreeName string) (int, string) {
	if al.UserConfig.RedisStrategy() == "database" {
		db := al.nextAvailableRedisDB()
		return db, ""
	}
	return 0, fmt.Sprintf("%s:%s", al.ProjectConfig.Project(), worktreeName)
}

func (al *Allocator) nextAvailableRedisDB() int {
	usedSet := make(map[int]bool)
	for _, db := range al.Registry.UsedRedisDbs() {
		usedSet[db] = true
	}
	for db := 1; db < maxRedisDbs; db++ {
		if !usedSet[db] {
			return db
		}
	}
	return 1
}

var sanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)
var collapseRe = regexp.MustCompile(`_+`)

func (al *Allocator) buildDatabaseName(worktreeName string) string {
	template := al.ProjectConfig.DatabaseTemplate()
	if template == "" {
		return ""
	}

	sanitized := sanitizeRe.ReplaceAllString(worktreeName, "_")
	sanitized = collapseRe.ReplaceAllString(sanitized, "_")
	sanitized = strings.Trim(sanitized, "_")

	return strings.NewReplacer(
		"{template}", template,
		"{worktree}", sanitized,
		"{project}", al.ProjectConfig.Project(),
	).Replace(al.ProjectConfig.DatabasePattern())
}

func intsToAny(ints []int) []any {
	result := make([]any, len(ints))
	for i, v := range ints {
		result[i] = v
	}
	return result
}
