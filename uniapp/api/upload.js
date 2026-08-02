import { http } from './request'
import { getFullApiBase } from '@/utils/api-base'

export const uploadApi = {
	// Returns { url: "http://..." } on success.
	image: (filePath, formData = {}) => {
		return new Promise((resolve, reject) => {
			const token = uni.getStorageSync('campus_gig_access_token')
			uni.uploadFile({
				url: getFullApiBase() + '/upload',
				filePath,
				name: 'file',
				formData,
				header: token ? { Authorization: 'Bearer ' + token } : {},
				success: (res) => {
					try {
						const body = JSON.parse(res.data)
						if (body.code === 0) resolve(body.data)
						else reject(new Error(body.message || 'upload failed'))
					} catch (e) { reject(e) }
				},
				fail: (e) => reject(new Error(e.errMsg || 'upload failed')),
			})
		})
	},
}
