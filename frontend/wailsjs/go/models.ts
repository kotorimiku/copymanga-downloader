export namespace main {
	
	export class BookInfo {
	    Series: string;
	    Author: string;
	    Description: string;
	    Genre: string;
	    Title: string;
	    Cover: string;
	
	    static createFrom(source: any = {}) {
	        return new BookInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Series = source["Series"];
	        this.Author = source["Author"];
	        this.Description = source["Description"];
	        this.Genre = source["Genre"];
	        this.Title = source["Title"];
	        this.Cover = source["Cover"];
	    }
	}
	export class ChapterInfo {
	    index: number;
	    uuid: string;
	    count: number;
	    size: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ChapterInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.uuid = source["uuid"];
	        this.count = source["count"];
	        this.size = source["size"];
	        this.name = source["name"];
	    }
	}
	export class Display {
	    value: number;
	    display: string;
	
	    static createFrom(source: any = {}) {
	        return new Display(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.display = source["display"];
	    }
	}
	export class PathWord {
	    name: string;
	    path_word: string;
	
	    static createFrom(source: any = {}) {
	        return new PathWord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path_word = source["path_word"];
	    }
	}
	export class Comic {
	    name: string;
	    uuid: string;
	    cover: string;
	    path_word: string;
	    author: PathWord[];
	    theme: PathWord[];
	    brief: string;
	    region: Display;
	
	    static createFrom(source: any = {}) {
	        return new Comic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.uuid = source["uuid"];
	        this.cover = source["cover"];
	        this.path_word = source["path_word"];
	        this.author = this.convertValues(source["author"], PathWord);
	        this.theme = this.convertValues(source["theme"], PathWord);
	        this.brief = source["brief"];
	        this.region = this.convertValues(source["region"], Display);
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
	export class User {
	    username: string;
	    password: string;
	    token: string;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	        this.token = source["token"];
	    }
	}
	export class Config {
	    urlBase: string;
	    outputPath: string;
	    packageType: string;
	    userList: User[];
	    namingStyle: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.urlBase = source["urlBase"];
	        this.outputPath = source["outputPath"];
	        this.packageType = source["packageType"];
	        this.userList = this.convertValues(source["userList"], User);
	        this.namingStyle = source["namingStyle"];
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
	
	export class DownloaderSingle {
	    pathWord: string;
	    chapter?: ChapterInfo;
	    bookInfo?: BookInfo;
	    progress: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloaderSingle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pathWord = source["pathWord"];
	        this.chapter = this.convertValues(source["chapter"], ChapterInfo);
	        this.bookInfo = this.convertValues(source["bookInfo"], BookInfo);
	        this.progress = source["progress"];
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

