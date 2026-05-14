export namespace main {
	
	export class ChatInfo {
	    id: string;
	    parent_chat_id?: string;
	    title: string;
	    message_count: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.parent_chat_id = source["parent_chat_id"];
	        this.title = source["title"];
	        this.message_count = source["message_count"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class CommandPreview {
	    command: string;
	    is_dangerous: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.is_dangerous = source["is_dangerous"];
	        this.message = source["message"];
	    }
	}
	export class MessageInfo {
	    id: string;
	    chat_id: string;
	    role: string;
	    content: string;
	    model?: string;
	    provider?: string;
	    is_summary: boolean;
	    created_at: string;
	    updated_at: string;
	    finished_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new MessageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.chat_id = source["chat_id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.model = source["model"];
	        this.provider = source["provider"];
	        this.is_summary = source["is_summary"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.finished_at = source["finished_at"];
	    }
	}

}

