import {ref} from 'vue'
import {request} from '@syl/base-request'

export const useResetPassword = () => {
    const loading = ref(false)

    async function resetPassword(data: {
        username?: string
        email?: string
        phone?: string
        new_password: string
    }) {
        loading.value = true
        try {
            return await request.post('user/reset-password', data, {
                showError: true,
                rethrow: true,
            })
        } finally {
            loading.value = false
        }
    }

    return {loading, resetPassword}
}