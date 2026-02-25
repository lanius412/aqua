package list

import (
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/fatih/color"
)

func (c *Controller) listInstalled(logger *slog.Logger, param *config.Param) error {
	cfgFilePaths := c.configFinder.Finds(param.CWD, param.ConfigFilePath)
	cfgFileMap := map[string]struct{}{}
	for _, cfgFilePath := range cfgFilePaths {
		if _, ok := cfgFileMap[cfgFilePath]; ok {
			continue
		}
		cfgFileMap[cfgFilePath] = struct{}{}

		if err := c.listInstalledByConfig(logger, cfgFilePath); err != nil {
			return slogerr.With(err, //nolint:wrapcheck
				"config_file_path", cfgFilePath,
			)
		}
	}

	if !param.All {
		return nil
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		logger := logger.With("config_file_path", cfgFilePath)
		if _, ok := cfgFileMap[cfgFilePath]; ok {
			continue
		}
		cfgFileMap[cfgFilePath] = struct{}{}

		logger.Debug("checking a global configuration file")
		if _, err := c.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := c.listInstalledByConfig(logger, cfgFilePath); err != nil {
			return slogerr.With(err, //nolint:wrapcheck
				"config_file_path", cfgFilePath,
			)
		}
	}
	return nil
}

func (c *Controller) listInstalledByConfig(logger *slog.Logger, cfgFilePath string) error {
	cfg := &aqua.Config{}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	table := tablewriter.NewTable(c.stdout,
		tablewriter.WithHeader([]string{"Package", "Version", "Registry"}),
		tablewriter.WithRenderer(
			renderer.NewColorized(
				renderer.ColorizedConfig{
					Header: renderer.Tint{
						FG: renderer.Colors{
							color.Italic,
							color.FgHiBlue,
						},
					},
					Column: renderer.Tint{
						FG: renderer.Colors{
							color.Reset,
						},
					},
				},
			),
		),
		tablewriter.WithRendition(
			tw.Rendition{
				Borders: tw.BorderNone,
				Settings: tw.Settings{
					Separators: tw.SeparatorsNone,
					Lines: tw.LinesNone,
				},
			},
		),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
	)

	data := make([][]string, len(cfg.Packages))
	for i, pkg := range cfg.Packages {
		data[i] = []string{
			pkg.Name,
			pkg.Version,
			pkg.Registry,
		}
	}

	table.Bulk(data)

	return table.Render()
}
