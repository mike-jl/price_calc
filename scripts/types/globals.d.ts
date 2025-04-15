import type Alpine from 'alpinejs';
import type {Htmx} from 'htmx.org';

declare global {
    interface Window {
        htmx: Htmx
        Alpine: typeof Alpine
    }
}
