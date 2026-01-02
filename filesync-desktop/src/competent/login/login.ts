import {request} from "@syl/base-request"

export const useLogin=()=>{
    async function login(data:any){
        return await request.post("auth/login", data,{
            showError:true,
            rethrow:true
        })
    }
    async function verify(){
        return await request.post("auth/verify",{},{
            showError:true,
            rethrow:true
        })
    }
    return {
        login,
        verify
    }
}