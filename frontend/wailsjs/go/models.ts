export namespace main {
	
	export class CommandPreview {
	    Command: string;
	    IsDangerous: boolean;
	    Message: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Command = source["Command"];
	        this.IsDangerous = source["IsDangerous"];
	        this.Message = source["Message"];
	    }
	}

}

