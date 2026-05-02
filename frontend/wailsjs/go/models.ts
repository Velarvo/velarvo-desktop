export namespace app {
	
	export class ExplorerNode {
	    id: string;
	    name: string;
	    path: string;
	    type: string;
	    size?: number;
	    modified: number;
	
	    static createFrom(source: any = {}) {
	        return new ExplorerNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	    }
	}

}

export namespace types {
	
	export class APIResponse_map_string_string_ {
	    success: boolean;
	    code: string;
	    message: string;
	    data?: Record<string, string>;
	    error?: string;
	    statusCode?: number;
	
	    static createFrom(source: any = {}) {
	        return new APIResponse_map_string_string_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = source["data"];
	        this.error = source["error"];
	        this.statusCode = source["statusCode"];
	    }
	}

}

