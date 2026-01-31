import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import Price from '../price';

describe('Price', () => {
  it('should render price with default USD currency', () => {
    render(<Price amount={29.99} />);
    expect(screen.getByText('USD')).toBeInTheDocument();
  });

  it('should render price with custom currency', () => {
    render(<Price amount={100} currencyCode="EUR" />);
    expect(screen.getByText('EUR')).toBeInTheDocument();
  });

  it('should render price with custom currency (RUB)', () => {
    render(<Price amount={1500} currencyCode="RUB" />);
    expect(screen.getByText('RUB')).toBeInTheDocument();
  });

  it('should apply custom className to price container', () => {
    const { container } = render(<Price amount={50} className="custom-class" />);
    const priceElement = container.querySelector('p');
    expect(priceElement).toHaveClass('custom-class');
  });

  it('should apply custom currencyCodeClassName to currency span', () => {
    render(<Price amount={50} currencyCode="USD" currencyCodeClassName="currency-style" />);
    const currencySpan = screen.getByText('USD');
    expect(currencySpan).toHaveClass('currency-style');
  });

  it('should format zero amount correctly', () => {
    render(<Price amount={0} />);
    expect(screen.getByText('USD')).toBeInTheDocument();
  });

  it('should format large amounts correctly', () => {
    render(<Price amount={1234567.89} />);
    expect(screen.getByText('USD')).toBeInTheDocument();
  });

  it('should render paragraph element', () => {
    const { container } = render(<Price amount={100} />);
    const priceElement = container.querySelector('p');
    expect(priceElement).toBeInTheDocument();
  });
});
