import type { FarmerProfile, GrainEntry, GrainPreset } from '@/types/grain'

export interface GrainSeedData {
  farmers: FarmerProfile[]
  entries: GrainEntry[]
  preset: GrainPreset
}

export function getLocalGrainSeed(): GrainSeedData {
  return {
    preset: {
      salesmanName: '王强',
      crops: ['小麦', '玉米', '大豆', '花生'],
      payTypes: ['银行卡', '现金', '对公转账'],
      places: ['北城一号仓', '丰收村临时收购点', '南环收购点'],
    },
    farmers: [
      {
        id: 'farmerLi',
        name: '李建国',
        idNumber: '410***********3215',
        phone: '138****5621',
        address: '河南省周口市示范区北城街道丰收村 3 组',
        bankNumber: '6228 **** **** 6631',
        bankName: '农商银行北城支行',
        status: 'complete',
        statusText: '资料完整',
      },
      {
        id: 'farmerZhang',
        name: '张秀兰',
        idNumber: '410***********2842',
        phone: '136****0927',
        address: '河南省周口市示范区南环街道向阳村 6 组',
        bankNumber: '待补充',
        bankName: '待补充',
        status: 'missing-bank',
        statusText: '银行卡照片待补',
      },
      {
        id: 'farmerChen',
        name: '陈玉山',
        idNumber: '410***********1178',
        phone: '139****6180',
        address: '河南省周口市示范区西郊街道粮丰村 1 组',
        bankNumber: '现金付款',
        bankName: '无',
        status: 'complete',
        statusText: '资料完整',
      },
    ],
    entries: [
      {
        id: 'entry-1',
        farmerId: 'farmerLi',
        crop: '小麦',
        quantity: 4200,
        unit: '斤',
        amount: 5460,
        buyTime: '2026-05-11 10:36',
        place: '北城一号仓',
        payType: '银行卡',
      },
      {
        id: 'entry-2',
        farmerId: 'farmerLi',
        crop: '小麦',
        quantity: 3600,
        unit: '斤',
        amount: 4680,
        buyTime: '2026-05-11 09:48',
        place: '北城一号仓',
        payType: '银行卡',
      },
      {
        id: 'entry-3',
        farmerId: 'farmerLi',
        crop: '小麦',
        quantity: 4800,
        unit: '斤',
        amount: 6240,
        buyTime: '2026-05-11 08:22',
        place: '丰收村临时收购点',
        payType: '银行卡',
      },
      {
        id: 'entry-4',
        farmerId: 'farmerZhang',
        crop: '玉米',
        quantity: 5000,
        unit: '斤',
        amount: 5800,
        buyTime: '2026-05-11 09:15',
        place: '南环收购点',
        payType: '银行卡',
      },
      {
        id: 'entry-5',
        farmerId: 'farmerZhang',
        crop: '玉米',
        quantity: 3400,
        unit: '斤',
        amount: 3944,
        buyTime: '2026-05-11 08:18',
        place: '南环收购点',
        payType: '银行卡',
      },
      {
        id: 'entry-6',
        farmerId: 'farmerChen',
        crop: '小麦',
        quantity: 5200,
        unit: '斤',
        amount: 6760,
        buyTime: '2026-05-11 08:40',
        place: '北城一号仓',
        payType: '现金',
      },
    ],
  }
}
