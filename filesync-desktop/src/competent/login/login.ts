import {request} from "@syl/base-request"

export const useLogin = () => {
    async function login(data: any) {
        return await request.post("user/login", data, {
            showError: true,
            rethrow: true
        })
    }

    async function verify() {
        return await request.post("user/verify", {}, {
            showError: true,
            rethrow: true
        })
    }

    return {
        login,
        verify
    }
}