package list

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/fatih/color"
)

func (c *Controller) List(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	if param.Installed {
		return c.listInstalled(logger, param)
	}
	cfg := &aqua.Config{}
	cfgFilePath, err := c.configFinder.Find(param.CWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	checksums, updateChecksum, err := checksum.Open(
		logger, c.fs, cfgFilePath,
		param.ChecksumEnabled(cfg))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logger, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}

	table := tablewriter.NewTable(c.stdout,
		tablewriter.WithHeader([]string{"Package", "Registry"}),
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

	data := make([][]string, len(registryContents))
	for registryName, registryContent := range registryContents {
		for pkgName := range registryContent.PackageInfos.ToMap(logger) {
			if pkgName == "" {
				logger.Debug("ignore a package because the package name is empty")
				continue
			}
			data = append(data, []string{pkgName, registryName})
		}
	}

	table.Bulk(data)

	return table.Render()
}
