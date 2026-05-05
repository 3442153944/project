import {ref} from 'vue'
import {request} from '@syl/base-request'

export const useRegister = () => {
    const loading = ref(false)

    async function register(data: {
        username?: string
        password: string
        email?: string
        phone?: string
        avatar?: string
    }) {
        loading.value = true
        try {
            return await request.post('user/register', data, {
                showError: true,
                rethrow: true,
            })
        } finally {
            loading.value = false
        }
    }

    return {
        loading,
        register,
    }
}