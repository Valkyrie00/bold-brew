package models

//type Formulae []Formula

type Formula struct {
	Name                    string        `json:"name"`
	FullName                string        `json:"full_name"`
	Tap                     string        `json:"tap"`
	OldNames                []string      `json:"oldnames"`
	Aliases                 []string      `json:"aliases"`
	VersionedFormulae       []string      `json:"versioned_formulae"`
	Description             string        `json:"desc"`
	License                 string        `json:"license"`
	Homepage                string        `json:"homepage"`
	Versions                Versions      `json:"versions"`
	Urls                    Urls          `json:"urls"`
	Revision                int           `json:"revision"`
	VersionScheme           int           `json:"version_scheme"`
	Bottle                  Bottle        `json:"bottle"`
	PourBottleOnlyIf        interface{}   `json:"pour_bottle_only_if"`
	KegOnly                 bool          `json:"keg_only"`
	KegOnlyReason           interface{}   `json:"keg_only_reason"`
	Options                 []interface{} `json:"options"`
	BuildDependencies       []string      `json:"build_dependencies"`
	Dependencies            []string      `json:"dependencies"`
	TestDependencies        []interface{} `json:"test_dependencies"`
	RecommendedDependencies []interface{} `json:"recommended_dependencies"`
	OptionalDependencies    []interface{} `json:"optional_dependencies"`
	//UsesFromMacOS           []string              `json:"uses_from_macos"`
	//UsesFromMacOSBounds     []UsesFromMacOSBounds `json:"uses_from_macos_bounds"`
	Requirements           []interface{}      `json:"requirements"`
	ConflictsWith          []interface{}      `json:"conflicts_with"`
	ConflictsWithReasons   []interface{}      `json:"conflicts_with_reasons"`
	LinkOverwrite          []interface{}      `json:"link_overwrite"`
	Caveats                interface{}        `json:"caveats"`
	Installed              []Installed        `json:"installed"`
	LinkedKeg              string             `json:"linked_keg"`
	Pinned                 bool               `json:"pinned"`
	Outdated               bool               `json:"outdated"`
	Deprecated             bool               `json:"deprecated"`
	DeprecationDate        interface{}        `json:"deprecation_date"`
	DeprecationReason      interface{}        `json:"deprecation_reason"`
	DeprecationReplacement interface{}        `json:"deprecation_replacement"`
	Disabled               bool               `json:"disabled"`
	DisableDate            interface{}        `json:"disable_date"`
	DisableReason          interface{}        `json:"disable_reason"`
	DisableReplacement     interface{}        `json:"disable_replacement"`
	PostInstallDefined     bool               `json:"post_install_defined"`
	Service                interface{}        `json:"service"`
	TapGitHead             string             `json:"tap_git_head"`
	RubySourcePath         string             `json:"ruby_source_path"`
	RubySourceChecksum     RubySourceChecksum `json:"ruby_source_checksum"`
}

type Versions struct {
	Stable string `json:"stable"`
	Head   string `json:"head"`
	Bottle bool   `json:"bottle"`
}

type Urls struct {
	Stable Url `json:"stable"`
	Head   Url `json:"head"`
}

type Url struct {
	Url      string      `json:"url"`
	Tag      interface{} `json:"tag"`
	Revision interface{} `json:"revision"`
	Using    interface{} `json:"using"`
	Checksum string      `json:"checksum"`
	Branch   string      `json:"branch"`
}

type Bottle struct {
	Stable BottleStable `json:"stable"`
}

type BottleStable struct {
	Rebuild int                   `json:"rebuild"`
	RootUrl string                `json:"root_url"`
	Files   map[string]BottleFile `json:"files"`
}

type BottleFile struct {
	Cellar string `json:"cellar"`
	Url    string `json:"url"`
	Sha256 string `json:"sha256"`
}

type UsesFromMacOSBounds struct {
}

type Installed struct {
	Version               string              `json:"version"`
	UsedOptions           []interface{}       `json:"used_options"`
	BuiltAsBottle         bool                `json:"built_as_bottle"`
	PouredFromBottle      bool                `json:"poured_from_bottle"`
	Time                  int64               `json:"time"`
	RuntimeDependencies   []RuntimeDependency `json:"runtime_dependencies"`
	InstalledAsDependency bool                `json:"installed_as_dependency"`
	InstalledOnRequest    bool                `json:"installed_on_request"`
}

type RuntimeDependency struct {
	FullName         string `json:"full_name"`
	Version          string `json:"version"`
	Revision         int    `json:"revision"`
	PkgVersion       string `json:"pkg_version"`
	DeclaredDirectly bool   `json:"declared_directly"`
}

type RubySourceChecksum struct {
	Sha256 string `json:"sha256"`
}
