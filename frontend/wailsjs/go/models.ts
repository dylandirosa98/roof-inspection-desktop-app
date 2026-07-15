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

