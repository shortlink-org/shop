package ports

// UnitOfWork manages transaction lifecycle.
// It does NOT know about repositories — only about transactions.
// Repositories detect transaction in context and participate automatically.
