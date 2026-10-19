// Categorias disponíveis para inscrição em promoções.
// Quando o main expuser as categorias (ex.: GET /categories), basta trocar esta lista por uma query.
export const PROMOTION_CATEGORIES = [
  { id: 'papelaria', label: 'Papelaria' },
  { id: 'escritorio', label: 'Escritório' },
  { id: 'escolar', label: 'Material escolar' },
  { id: 'informatica', label: 'Informática' },
  { id: 'arte', label: 'Arte e desenho' },
] as const
