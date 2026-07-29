export namespace analysis {
	
	export class Detection {
	    Class: string;
	    Confidence: number;
	    X: number;
	    Y: number;
	    Width: number;
	    Height: number;
	
	    static createFrom(source: any = {}) {
	        return new Detection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Class = source["Class"];
	        this.Confidence = source["Confidence"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.Width = source["Width"];
	        this.Height = source["Height"];
	    }
	}
	export class BoundingBox {
	    Top: number;
	    Left: number;
	    Right: number;
	    Bottom: number;
	
	    static createFrom(source: any = {}) {
	        return new BoundingBox(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Top = source["Top"];
	        this.Left = source["Left"];
	        this.Right = source["Right"];
	        this.Bottom = source["Bottom"];
	    }
	}
	export class ImageDetection {
	    BoundingBox: BoundingBox;
	    Detection: Detection;
	
	    static createFrom(source: any = {}) {
	        return new ImageDetection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BoundingBox = this.convertValues(source["BoundingBox"], BoundingBox);
	        this.Detection = this.convertValues(source["Detection"], Detection);
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
	export class AnalysisResult {
	    OriginalImageBoxes: ImageDetection[];
	    ModelImageBoxes: ImageDetection[];
	
	    static createFrom(source: any = {}) {
	        return new AnalysisResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OriginalImageBoxes = this.convertValues(source["OriginalImageBoxes"], ImageDetection);
	        this.ModelImageBoxes = this.convertValues(source["ModelImageBoxes"], ImageDetection);
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

export namespace database {
	
	export class Project {
	    ID: number;
	    Name: string;
	    Directory: string;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Directory = source["Directory"];
	    }
	}
	export class RetrieveAiImagesRow {
	    ImageID: number;
	    ImagePath: string;
	    ImageWidth: sql.NullInt64;
	    ImageHeight: sql.NullInt64;
	    ImagePreviewUrl: sql.NullString;
	    AiImageID: sql.NullInt64;
	    AnnotationsJson: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new RetrieveAiImagesRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ImageID = source["ImageID"];
	        this.ImagePath = source["ImagePath"];
	        this.ImageWidth = this.convertValues(source["ImageWidth"], sql.NullInt64);
	        this.ImageHeight = this.convertValues(source["ImageHeight"], sql.NullInt64);
	        this.ImagePreviewUrl = this.convertValues(source["ImagePreviewUrl"], sql.NullString);
	        this.AiImageID = this.convertValues(source["AiImageID"], sql.NullInt64);
	        this.AnnotationsJson = this.convertValues(source["AnnotationsJson"], sql.NullString);
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
	export class RetrieveImagesRow {
	    Width: sql.NullInt64;
	    Height: sql.NullInt64;
	    FileSize: sql.NullInt64;
	    Format: sql.NullString;
	    Path: string;
	    DataUrl: sql.NullString;
	    PreviewUrl: sql.NullString;
	    ID: number;
	    ProjectID: number;
	
	    static createFrom(source: any = {}) {
	        return new RetrieveImagesRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Width = this.convertValues(source["Width"], sql.NullInt64);
	        this.Height = this.convertValues(source["Height"], sql.NullInt64);
	        this.FileSize = this.convertValues(source["FileSize"], sql.NullInt64);
	        this.Format = this.convertValues(source["Format"], sql.NullString);
	        this.Path = source["Path"];
	        this.DataUrl = this.convertValues(source["DataUrl"], sql.NullString);
	        this.PreviewUrl = this.convertValues(source["PreviewUrl"], sql.NullString);
	        this.ID = source["ID"];
	        this.ProjectID = source["ProjectID"];
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
	export class RetrieveProjectsRow {
	    ID: number;
	    Name: string;
	    Directory: string;
	    ImageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new RetrieveProjectsRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Directory = source["Directory"];
	        this.ImageCount = source["ImageCount"];
	    }
	}

}

export namespace sql {
	
	export class NullInt64 {
	    Int64: number;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullInt64(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Int64 = source["Int64"];
	        this.Valid = source["Valid"];
	    }
	}
	export class NullString {
	    String: string;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullString(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.String = source["String"];
	        this.Valid = source["Valid"];
	    }
	}

}

