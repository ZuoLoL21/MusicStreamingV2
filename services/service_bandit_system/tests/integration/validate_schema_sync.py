#!/usr/bin/env python3
"""
Validate that the test table schema matches the production view columns.

Run this after modifying database_data_warehouse/12_bandit_input.sql
to ensure the test schema stays in sync.
"""

import re
import sys
from pathlib import Path

# File paths
PROD_VIEW = Path(__file__).parent.parent.parent.parent / "database_data_warehouse" / "12_bandit_input.sql"
TEST_TABLE = Path(__file__).parent / "sql" / "init_clickhouse.sql"


def extract_view_columns(view_file: Path) -> set[str]:
    """Extract column names from the production view's SELECT clause."""
    content = view_file.read_text()

    # Find all "AS f_<name>" patterns in the SELECT clause
    column_pattern = r'\bAS\s+(f_[a-z_]+)'
    columns = set(re.findall(column_pattern, content, re.IGNORECASE))

    # Add the key columns
    columns.add('user_uuid')
    columns.add('theme')

    return columns


def extract_table_columns(table_file: Path) -> set[str]:
    """Extract column names from the test table definition."""
    content = table_file.read_text()

    # Find all column definitions (name followed by type)
    column_pattern = r'^\s+([a-z_]+)\s+(?:UUID|String|Float64)'
    columns = set(re.findall(column_pattern, content, re.MULTILINE | re.IGNORECASE))

    return columns


def main():
    if not PROD_VIEW.exists():
        print(f"❌ Production view not found: {PROD_VIEW}")
        return 1

    if not TEST_TABLE.exists():
        print(f"❌ Test table not found: {TEST_TABLE}")
        return 1

    view_cols = extract_view_columns(PROD_VIEW)
    table_cols = extract_table_columns(TEST_TABLE)

    # Check for mismatches
    missing_in_test = view_cols - table_cols
    extra_in_test = table_cols - view_cols

    if not missing_in_test and not extra_in_test:
        print(f"[OK] Schemas are in sync ({len(view_cols)} columns)")
        print(f"   View:  {PROD_VIEW}")
        print(f"   Table: {TEST_TABLE}")
        return 0

    print("[ERROR] Schema mismatch detected!")
    print(f"   View:  {PROD_VIEW}")
    print(f"   Table: {TEST_TABLE}")
    print()

    if missing_in_test:
        print(f"Missing in test table ({len(missing_in_test)}):")
        for col in sorted(missing_in_test):
            print(f"  - {col}")
        print()

    if extra_in_test:
        print(f"Extra in test table ({len(extra_in_test)}):")
        for col in sorted(extra_in_test):
            print(f"  + {col}")
        print()

    print("Action required:")
    print(f"  Update {TEST_TABLE.name} to match the production view")

    return 1


if __name__ == "__main__":
    sys.exit(main())
