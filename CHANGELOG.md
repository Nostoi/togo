# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed
- **Terminal resize crash fix**: Fixed critical crash that occurred during terminal orientation changes (e.g., iPhone portrait → landscape → portrait). The issue was caused by column/row count mismatches during table rendering when `WindowSizeMsg` events triggered table updates.
  - Root cause: `SetColumns()` internally called `UpdateViewport()` before `SetRows()` was called, creating temporary states where column count didn't match row cell counts
  - Solution: Completely rebuild the table with consistent columns and rows to prevent any intermediate inconsistent states
  - Added `normalizeCells()` helper function to ensure rows always have the correct number of cells
  - Added comprehensive resize tests to prevent regression
  - Added defensive checks for minimum column widths

### Added
- **Comprehensive resize testing**: Added test scenarios covering rapid resize events, column/row consistency validation, and extreme width scenarios
- **Defensive width handling**: Added minimum width constraints for all table columns to prevent zero/negative width issues

## Previous Changes
- (Previous changelog entries would go here)