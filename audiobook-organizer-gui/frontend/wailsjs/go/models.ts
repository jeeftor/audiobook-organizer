export namespace main {

	export class AlbumGroup {
	    name: string;
	    author: string;
	    series: string;
	    file_count: number;
	    file_indices: number[];
	    files: organizer.Metadata[];

	    static createFrom(source: any = {}) {
	        return new AlbumGroup(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.author = source["author"];
	        this.series = source["series"];
	        this.file_count = source["file_count"];
	        this.file_indices = source["file_indices"];
	        this.files = this.convertValues(source["files"], organizer.Metadata);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FieldMappingOption {
	    field: string;
	    label: string;
	    description: string;
	    options: string[];
	    current: string;

	    static createFrom(source: any = {}) {
	        return new FieldMappingOption(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.field = source["field"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.options = source["options"];
	        this.current = source["current"];
	    }
	}
	export class FieldMappingPreset {
	    name: string;
	    description: string;
	    mapping: organizer.FieldMapping;

	    static createFrom(source: any = {}) {
	        return new FieldMappingPreset(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.mapping = this.convertValues(source["mapping"], organizer.FieldMapping);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileOperation {
	    from: string;
	    to: string;

	    static createFrom(source: any = {}) {
	        return new FileOperation(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	    }
	}
	export class InitialDirectories {
	    input_dir: string;
	    output_dir: string;

	    static createFrom(source: any = {}) {
	        return new InitialDirectories(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input_dir = source["input_dir"];
	        this.output_dir = source["output_dir"];
	    }
	}
	export class LayoutOption {
	    name: string;
	    description: string;

	    static createFrom(source: any = {}) {
	        return new LayoutOption(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class RawFieldPreview {
	    key: string;
	    value: string;
	    indicator: string;

	    static createFrom(source: any = {}) {
	        return new RawFieldPreview(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	        this.indicator = source["indicator"];
	    }
	}
	export class MetadataPreview {
	    filename: string;
	    source_type: string;
	    raw_fields: RawFieldPreview[];
	    mapping: organizer.FieldMapping;

	    static createFrom(source: any = {}) {
	        return new MetadataPreview(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.source_type = source["source_type"];
	        this.raw_fields = this.convertValues(source["raw_fields"], RawFieldPreview);
	        this.mapping = this.convertValues(source["mapping"], organizer.FieldMapping);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PreviewItem {
	    from: string;
	    to: string;
	    is_conflict: boolean;
	    author: string;
	    series: string;
	    title: string;
	    filename: string;
	    output_dir: string;

	    static createFrom(source: any = {}) {
	        return new PreviewItem(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.is_conflict = source["is_conflict"];
	        this.author = source["author"];
	        this.series = source["series"];
	        this.title = source["title"];
	        this.filename = source["filename"];
	        this.output_dir = source["output_dir"];
	    }
	}
	export class ProgressUpdate {
	    status: string;
	    current: number;
	    total: number;
	    current_file: string;

	    static createFrom(source: any = {}) {
	        return new ProgressUpdate(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.current = source["current"];
	        this.total = source["total"];
	        this.current_file = source["current_file"];
	    }
	}

	export class ValidationWarning {
	    book_index: number;
	    book_title: string;
	    type: string;
	    message: string;
	    severity: string;

	    static createFrom(source: any = {}) {
	        return new ValidationWarning(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.book_index = source["book_index"];
	        this.book_title = source["book_title"];
	        this.type = source["type"];
	        this.message = source["message"];
	        this.severity = source["severity"];
	    }
	}

	export class RenameConfig {
	    enabled: boolean;
	    template: string;
	    preset: string;
	    separator: string;
	    author_format: string;
	    replace_spaces: boolean;
	    space_char: string;

	    static createFrom(source: any = {}) {
	        return new RenameConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.template = source["template"];
	        this.preset = source["preset"];
	        this.separator = source["separator"];
	        this.author_format = source["author_format"];
	        this.replace_spaces = source["replace_spaces"];
	        this.space_char = source["space_char"];
	    }
	}
	export class ScanMode {
	    name: string;
	    use_embedded_metadata: boolean;
	    flat: boolean;
	    description: string;

	    static createFrom(source: any = {}) {
	        return new ScanMode(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.use_embedded_metadata = source["use_embedded_metadata"];
	        this.flat = source["flat"];
	        this.description = source["description"];
	    }
	}
	export class ScanStatistics {
	    total_files: number;
	    total_audiobooks: number;
	    missing_metadata: number;
	    album_groups: AlbumGroup[];
	    ungrouped_files: organizer.Metadata[];

	    static createFrom(source: any = {}) {
	        return new ScanStatistics(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_files = source["total_files"];
	        this.total_audiobooks = source["total_audiobooks"];
	        this.missing_metadata = source["missing_metadata"];
	        this.album_groups = this.convertValues(source["album_groups"], AlbumGroup);
	        this.ungrouped_files = this.convertValues(source["ungrouped_files"], organizer.Metadata);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace organizer {

	export class FieldMapping {
	    title_field?: string;
	    series_field?: string;
	    author_fields?: string[];
	    track_field?: string;
	    disc_field?: string;

	    static createFrom(source: any = {}) {
	        return new FieldMapping(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title_field = source["title_field"];
	        this.series_field = source["series_field"];
	        this.author_fields = source["author_fields"];
	        this.track_field = source["track_field"];
	        this.disc_field = source["disc_field"];
	    }
	}
	export class Metadata {
	    title: string;
	    authors: string[];
	    series: string[];
	    track_number?: number;
	    album?: string;
	    track_title?: string;
	    source_type: string;
	    source_path: string;
	    raw_data?: Record<string, any>;

	    static createFrom(source: any = {}) {
	        return new Metadata(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.authors = source["authors"];
	        this.series = source["series"];
	        this.track_number = source["track_number"];
	        this.album = source["album"];
	        this.track_title = source["track_title"];
	        this.source_type = source["source_type"];
	        this.source_path = source["source_path"];
	        this.raw_data = source["raw_data"];
	    }
	}
	export class MoveSummary {
	    from: string;
	    to: string;

	    static createFrom(source: any = {}) {
	        return new MoveSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	    }
	}
	export class OrganizerConfig {
	    BaseDir: string;
	    OutputDir: string;
	    ReplaceSpace: string;
	    Verbose: boolean;
	    DryRun: boolean;
	    Undo: boolean;
	    Prompt: boolean;
	    RemoveEmpty: boolean;
	    UseEmbeddedMetadata: boolean;
	    Flat: boolean;
	    SkipErrors: boolean;
	    Layout: string;
	    AuthorFormat: string;
	    FieldMapping: FieldMapping;
	    AllowedSourcePaths: string[];

	    static createFrom(source: any = {}) {
	        return new OrganizerConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BaseDir = source["BaseDir"];
	        this.OutputDir = source["OutputDir"];
	        this.ReplaceSpace = source["ReplaceSpace"];
	        this.Verbose = source["Verbose"];
	        this.DryRun = source["DryRun"];
	        this.Undo = source["Undo"];
	        this.Prompt = source["Prompt"];
	        this.RemoveEmpty = source["RemoveEmpty"];
	        this.UseEmbeddedMetadata = source["UseEmbeddedMetadata"];
	        this.Flat = source["Flat"];
	        this.SkipErrors = source["SkipErrors"];
	        this.Layout = source["Layout"];
	        this.AuthorFormat = source["AuthorFormat"];
	        this.FieldMapping = this.convertValues(source["FieldMapping"], FieldMapping);
	        this.AllowedSourcePaths = source["AllowedSourcePaths"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Summary {
	    MetadataFound: string[];
	    MetadataMissing: string[];
	    Moves: MoveSummary[];
	    EmptyDirsRemoved: string[];

	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.MetadataFound = source["MetadataFound"];
	        this.MetadataMissing = source["MetadataMissing"];
	        this.Moves = this.convertValues(source["Moves"], MoveSummary);
	        this.EmptyDirsRemoved = source["EmptyDirsRemoved"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}
