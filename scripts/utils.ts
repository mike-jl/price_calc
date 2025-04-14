import { EditableWithId } from './types/common';

export function createEditingHelpers<T extends EditableWithId>(
    list: T[],
    backup: Record<number, T>
) {
    return {
        startEditing(item: T): void {
            backup[item.id] = JSON.parse(JSON.stringify(item)) as T;
            item.editing = true;
        },
        cancelEditing(item: T): void {
            const saved = backup[item.id];
            if (saved) {
                Object.assign(item, saved);
                delete backup[item.id];
            }
            item.editing = false;
        },
        removeItem(id: number): void {
            const index = list.findIndex((item) => item.id === id);
            if (index !== -1) list.splice(index, 1);
        }
    };
}

