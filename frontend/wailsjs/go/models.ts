export namespace inspection {
	
	export class Image {
	    Width: number;
	    Height: number;
	    Format: string;
	    FileSize: number;
	    Path: string;
	    DataURL: string;
	    PreviewURL: string;
	
	    static createFrom(source: any = {}) {
	        return new Image(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Width = source["Width"];
	        this.Height = source["Height"];
	        this.Format = source["Format"];
	        this.FileSize = source["FileSize"];
	        this.Path = source["Path"];
	        this.DataURL = source["DataURL"];
	        this.PreviewURL = source["PreviewURL"];
	    }
	}
	export class ProjectImage {
	    ID: number;
	    ProjectID: number;
	    Image?: Image;
	
	    static createFrom(source: any = {}) {
	        return new ProjectImage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.ProjectID = source["ProjectID"];
	        this.Image = this.convertValues(source["Image"], Image);
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
	export class Project {
	    ID: number;
	    Directory: string;
	    Name: string;
	    Images: ProjectImage[];
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Directory = source["Directory"];
	        this.Name = source["Name"];
	        this.Images = this.convertValues(source["Images"], ProjectImage);
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

