package catalog

import (
	"fmt"
)

func Create(name string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Catalogs[name]; ok {
		return fmt.Errorf("catalog %q already exists", name)
	}
	cfg.Catalogs[name] = Catalog{DisplayName: name}
	if err := WriteConfig(cfg); err != nil {
		return err
	}
	content, err := setCatalogMetaData([]byte{}, MetaData{Name: name, DisplayName: name})
	if err != nil {
		return err
	}
	if err := WriteCatalogFile(name, content); err != nil {
		return err
	}
	fmt.Printf("created empty catalog %s\n", name)
	return nil
}
