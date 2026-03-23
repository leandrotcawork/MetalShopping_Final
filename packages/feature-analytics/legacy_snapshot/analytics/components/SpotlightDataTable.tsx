import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  type ColumnDef,
  type SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { useMemo, useState } from "react";
import type { ReactNode } from "react";

import styles from "./spotlight_data_table.module.css";

type ColumnAlign = "left" | "right" | "center";

export type SpotlightDataTableColumn<T> = {
  id: string;
  header: string;
  accessor: (row: T) => string | number | null | undefined;
  cell?: (row: T) => ReactNode;
  sortValue?: (row: T) => string | number | null | undefined;
  align?: ColumnAlign;
};

type SpotlightDataTableProps<T> = {
  rows: T[];
  columns: SpotlightDataTableColumn<T>[];
  emptyText: string;
  rowKey: (row: T, index: number) => string;
  defaultSort?: { id: string; desc?: boolean };
};

function normalizeSortValue(value: string | number | null | undefined): string | number {
  if (typeof value === "number") return Number.isFinite(value) ? value : Number.NEGATIVE_INFINITY;
  return String(value ?? "").toLocaleLowerCase("pt-BR");
}

function alignClassName(align: ColumnAlign | undefined): string {
  if (align === "right") return styles.alignRight;
  if (align === "center") return styles.alignCenter;
  return "";
}

export function SpotlightDataTable<T>({
  rows,
  columns,
  emptyText,
  rowKey,
  defaultSort,
}: SpotlightDataTableProps<T>) {
  const [sorting, setSorting] = useState<SortingState>(
    defaultSort ? [{ id: defaultSort.id, desc: Boolean(defaultSort.desc) }] : []
  );

  const tableColumns = useMemo<ColumnDef<T, unknown>[]>(
    () =>
      columns.map((column) => ({
        id: column.id,
        accessorFn: (row) => column.accessor(row),
        header: ({ column: instance }) => {
          const sortState = instance.getIsSorted();
          const indicator = sortState === "asc" ? "▲" : sortState === "desc" ? "▼" : "↕";
          return (
            <button
              type="button"
              className={styles.sortButton}
              onClick={() => instance.toggleSorting(sortState === "asc")}
            >
              <span>{column.header}</span>
              <span className={styles.sortIndicator}>{indicator}</span>
            </button>
          );
        },
        cell: ({ row }) =>
          column.cell ? column.cell(row.original) : String(column.accessor(row.original) ?? "-"),
        sortingFn: (rowA, rowB) => {
          const valueA = normalizeSortValue(
            column.sortValue ? column.sortValue(rowA.original) : column.accessor(rowA.original)
          );
          const valueB = normalizeSortValue(
            column.sortValue ? column.sortValue(rowB.original) : column.accessor(rowB.original)
          );
          if (typeof valueA === "number" && typeof valueB === "number") return valueA - valueB;
          return String(valueA).localeCompare(String(valueB), "pt-BR");
        },
        meta: { align: column.align || "left" },
      })),
    [columns]
  );

  const table = useReactTable({
    data: rows,
    columns: tableColumns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getRowId: (row, index) => rowKey(row, index),
  });

  return (
    <>
      <table className={styles.table}>
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                const align = alignClassName((header.column.columnDef.meta as { align?: ColumnAlign } | undefined)?.align);
                return (
                  <th key={header.id} className={`${styles.th} ${align}`.trim()}>
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                );
              })}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id}>
              {row.getVisibleCells().map((cell) => {
                const align = alignClassName((cell.column.columnDef.meta as { align?: ColumnAlign } | undefined)?.align);
                return (
                  <td key={cell.id} className={`${styles.td} ${align}`.trim()}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
      {!rows.length ? <p className={styles.empty}>{emptyText}</p> : null}
    </>
  );
}
